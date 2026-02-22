package runenv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/workspace"
)

func LoadDecryptedEnv(envFilePath, wsRoot string) (map[string]string, error) {
	envFile, err := envfile.Load(envFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	masterKey, err := GetMasterKeyForWorkspace(wsRoot)
	if err != nil {
		return nil, err
	}

	env := crypto.NewEnvelope(masterKey)
	decrypted, err := envFile.DecryptAll(env)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}

func LoadDecryptedEnvFromFiles(paths []string, overload, strict bool) (map[string]string, error) {
	merged := make(map[string]string)
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("resolve path %q: %w", p, err)
		}
		if strict {
			if _, err := os.Stat(absPath); err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("missing env file: %s", absPath)
				}
				return nil, fmt.Errorf("env file %s: %w", absPath, err)
			}
		}

		wsRoot, err := workspace.FindRoot(filepath.Dir(absPath))
		if err != nil {
			if strict {
				return nil, fmt.Errorf("find workspace for %s: %w", absPath, err)
			}
			continue
		}

		decrypted, err := LoadDecryptedEnv(absPath, wsRoot)
		if err != nil {
			if strict {
				return nil, err
			}
			continue
		}
		for k, v := range decrypted {
			if overload || merged[k] == "" {
				merged[k] = v
			}
		}
	}
	return merged, nil
}

func MergeOverlayEnv(env map[string]string, overlay []string, overload bool) error {
	for _, s := range overlay {
		idx := strings.Index(s, "=")
		if idx <= 0 {
			return fmt.Errorf("invalid --env %q: expected KEY=value", s)
		}
		key := s[:idx]
		value := s[idx+1:]
		if overload || env[key] == "" {
			env[key] = value
		}
	}
	return nil
}

func buildCmdEnv(envMap map[string]string) []string {
	cmdEnv := os.Environ()
	for k, v := range envMap {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return cmdEnv
}

func setupCommand(command string, args []string, envMap map[string]string, workdir string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	cmd.Env = buildCmdEnv(envMap)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if workdir != "" {
		cmd.Dir = workdir
	}
	// Do not set Setpgid: child stays in our process group so Ctrl+C kills it too.
	return cmd
}

func exitCodeFromError(runErr error) (int, error) {
	if runErr == nil {
		return 0, nil
	}
	if exitErr, ok := runErr.(*exec.ExitError); ok {
		return exitErr.ExitCode(), runErr
	}
	return -1, fmt.Errorf("failed to run command: %w", runErr)
}

func RunWithEnvFromMap(envMap map[string]string, workdir, command string, args []string) (int, error) {
	cmd := setupCommand(command, args, envMap, workdir)
	return exitCodeFromError(cmd.Run())
}

func RunWithEnvRedactedFromMap(envMap map[string]string, workdir, command string, args []string) (int, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = buildCmdEnv(envMap)
	cmd.Stdin = os.Stdin
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if workdir != "" {
		cmd.Dir = workdir
	}
	// Do not set Setpgid: child stays in our process group so Ctrl+C kills it too.

	runErr := cmd.Run()
	stdoutBytes := redactOutput([]byte(stdout.String()), envMap)
	stderrBytes := redactOutput([]byte(stderr.String()), envMap)

	_, _ = os.Stdout.Write(stdoutBytes)
	_, _ = os.Stderr.Write(stderrBytes)

	return exitCodeFromError(runErr)
}

func redactOutput(data []byte, envMap map[string]string) []byte {
	result := string(data)
	for k, v := range envMap {
		if v != "" {
			result = strings.ReplaceAll(result, v, fmt.Sprintf("[REDACTED:%s]", k))
		}
	}
	return []byte(result)
}

func GetMasterKeyForWorkspace(wsRoot string) (*crypto.MasterKey, error) {
	if !workspace.WorkspaceFileExists(wsRoot) {
		return nil, fmt.Errorf("workspace not initialized - run 'openenvx init' first")
	}

	wc, err := workspace.ReadWorkspaceFile(wsRoot)
	if err != nil {
		return nil, fmt.Errorf("read workspace config: %w", err)
	}
	if wc.PublicKey == "" {
		return nil, fmt.Errorf("workspace missing public key - may be corrupted")
	}

	identity, ok := crypto.GetPrivateKey(wsRoot)
	if !ok {
		return nil, fmt.Errorf("private key not found - set OPENENVX_PRIVATE_KEY or add key with 'openenvx key add'")
	}

	strategy := crypto.NewAsymmetricStrategy(identity)
	return strategy.GetMasterKey()
}

func ResolveEnvPath(path, workdir string) (string, error) {
	if path == "" {
		path = ".env"
	}
	if workdir != "" {
		path = filepath.Join(workdir, path)
	}
	return filepath.Abs(path)
}

const MaxEnvSearchDepth = 16

func FindEnvInParents(dir string, maxDepth int) (string, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve directory: %w", err)
	}
	for i := 0; i < maxDepth; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .env found in current or parent directories (searched up to %d levels)", maxDepth)
}

func GetEnvelopes(envPath string, keys []string) (publicKey string, entries map[string]string, err error) {
	wsRoot, err := workspace.FindRoot(filepath.Dir(envPath))
	if err != nil {
		return "", nil, fmt.Errorf("find workspace: %w", err)
	}

	wc, err := workspace.ReadWorkspaceFile(wsRoot)
	if err != nil {
		return "", nil, fmt.Errorf("read workspace config: %w", err)
	}
	if wc.PublicKey == "" {
		return "", nil, fmt.Errorf("workspace missing public key")
	}

	envFile, err := envfile.Load(envPath)
	if err != nil {
		return "", nil, fmt.Errorf("load .env: %w", err)
	}

	entries = make(map[string]string)
	wantAll := len(keys) == 0
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, key := range envFile.Keys() {
		if !wantAll && !keySet[key] {
			continue
		}
		val, ok := envFile.Get(key)
		if !ok {
			continue
		}
		if _, err := crypto.ParseEncryptedValue(val); err != nil {
			continue
		}
		entries[key] = val
	}
	return wc.PublicKey, entries, nil
}

func ListEncryptedKeys(envPath string) ([]string, error) {
	_, entries, err := GetEnvelopes(envPath, nil)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func MaskSecretValue(value string) string {
	length := len(value)
	if length == 0 {
		return ""
	}
	switch {
	case length <= 4:
		return strings.Repeat("*", length)
	case length <= 8:
		return strings.Repeat("*", length-2) + value[length-2:]
	default:
		return strings.Repeat("*", length-4) + value[length-4:]
	}
}

func GetMaskedSecret(envPath, key string) (masked string, valueLength int, err error) {
	wsRoot, err := workspace.FindRoot(filepath.Dir(envPath))
	if err != nil {
		return "", 0, fmt.Errorf("find workspace: %w", err)
	}

	decrypted, err := LoadDecryptedEnv(envPath, wsRoot)
	if err != nil {
		return "", 0, err
	}
	value, ok := decrypted[key]
	if !ok {
		return "", 0, fmt.Errorf("key %q not found or not encrypted", key)
	}
	return MaskSecretValue(value), len(value), nil
}

func SetSecret(envPath, key, value string) error {
	wsRoot, err := workspace.FindRoot(filepath.Dir(envPath))
	if err != nil {
		return fmt.Errorf("find workspace: %w", err)
	}

	envFile, err := envfile.Load(envPath)
	if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}
	masterKey, err := GetMasterKeyForWorkspace(wsRoot)
	if err != nil {
		return err
	}
	env := crypto.NewEnvelope(masterKey)
	encrypted, err := env.Encrypt([]byte(value), key)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}
	envFile.Set(key, encrypted.String())
	if err := envFile.Save(); err != nil {
		return fmt.Errorf("save .env: %w", err)
	}
	return nil
}

func DeleteSecret(envPath, key string) (bool, error) {
	envFile, err := envfile.Load(envPath)
	if err != nil {
		return false, fmt.Errorf("load .env: %w", err)
	}
	if !envFile.Delete(key) {
		return false, nil
	}
	if err := envFile.Save(); err != nil {
		return false, fmt.Errorf("save .env: %w", err)
	}
	return true, nil
}

func KeyExists(envPath, key string) (bool, error) {
	_, entries, err := GetEnvelopes(envPath, []string{key})
	if err != nil {
		return false, err
	}
	_, exists := entries[key]
	return exists, nil
}

type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func RunWithEnvCaptured(envMap map[string]string, workdir, command string, args []string) (*RunResult, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = buildCmdEnv(envMap)
	cmd.Stdin = os.Stdin
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if workdir != "" {
		cmd.Dir = workdir
	}
	// Do not set Setpgid: child stays in our process group so Ctrl+C kills it too.

	runErr := cmd.Run()
	result := &RunResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	result.ExitCode, _ = exitCodeFromError(runErr)
	if runErr != nil && result.ExitCode == -1 {
		return result, runErr
	}
	return result, runErr
}

func RunWithEnvCapturedRedacted(envMap map[string]string, workdir, command string, args []string) (*RunResult, error) {
	result, err := RunWithEnvCaptured(envMap, workdir, command, args)
	if result != nil {
		result.Stdout = string(redactOutput([]byte(result.Stdout), envMap))
		result.Stderr = string(redactOutput([]byte(result.Stderr), envMap))
	}
	return result, err
}

var DevServerCommands = map[string]bool{
	"next":            true,
	"vite":            true,
	"ng":              true,
	"vue-cli-service": true,
	"react-scripts":   true,
	"wrangler":        true,
	"serve":           true,
	"nodemon":         true,
}

func IsDevServerCommand(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	base := filepath.Base(parts[0])
	if strings.HasSuffix(base, ".exe") {
		base = base[:len(base)-4]
	}

	if DevServerCommands[base] {
		return true
	}

	if base == "node" && len(parts) > 1 {
		return true
	}

	if base == "npm" || base == "yarn" || base == "pnpm" || base == "bun" {
		return true
	}

	return false
}

type ProcessRunner struct {
	Command string
	Args    []string
	Env     map[string]string
	Workdir string
	Redact  bool

	cmd       *exec.Cmd
	redactors []*streamRedactor
}

type streamRedactor struct {
	envMap map[string]string
	dst    *os.File
	pipeR  *os.File
}

func (sr *streamRedactor) run() {
	buf := make([]byte, 4096)
	for {
		n, err := sr.pipeR.Read(buf)
		if n > 0 {
			data := redactOutput(buf[:n], sr.envMap)
			sr.dst.Write(data)
		}
		if err != nil {
			return
		}
	}
}

func (r *ProcessRunner) Start() error {
	r.cmd = exec.Command(r.Command, r.Args...)
	r.cmd.Env = buildCmdEnv(r.Env)
	r.cmd.Stdin = os.Stdin
	if r.Workdir != "" {
		r.cmd.Dir = r.Workdir
	}
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if r.Redact {
		stdoutR, stdoutW, _ := os.Pipe()
		stderrR, stderrW, _ := os.Pipe()
		r.cmd.Stdout = stdoutW
		r.cmd.Stderr = stderrW

		r.redactors = []*streamRedactor{
			{envMap: r.Env, dst: os.Stdout, pipeR: stdoutR},
			{envMap: r.Env, dst: os.Stderr, pipeR: stderrR},
		}
		for _, sr := range r.redactors {
			go sr.run()
		}
	} else {
		r.cmd.Stdout = os.Stdout
		r.cmd.Stderr = os.Stderr
	}

	return r.cmd.Start()
}

var execPgrep = func(ppid int) ([]byte, error) {
	return exec.Command("pgrep", "-P", fmt.Sprintf("%d", ppid)).Output()
}

// killFunc is injectable for tests; production uses syscall.Kill.
var killFunc = func(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}

func getChildPids(rootPgid int) ([]int, error) {
	out, err := execPgrep(rootPgid)
	if err != nil {
		return nil, err
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var pid int
		if _, err := fmt.Sscanf(line, "%d", &pid); err == nil && pid > 0 {
			pids = append(pids, pid)
			children, _ := getChildPids(pid)
			pids = append(pids, children...)
		}
	}
	return pids, nil
}

func killProcessTree(pgid int) error {
	pids, err := getChildPids(pgid)
	if err == nil {
		for _, pid := range pids {
			_ = killFunc(pid, syscall.SIGTERM)
		}
	}
	return killFunc(-pgid, syscall.SIGTERM)
}

func (r *ProcessRunner) Stop() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}
	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err != nil {
		return r.cmd.Process.Kill()
	}
	if err := killProcessTree(pgid); err != nil {
		return r.cmd.Process.Kill()
	}
	done := make(chan error, 1)
	go func() {
		done <- r.cmd.Wait()
	}()
	select {
	case <-time.After(5 * time.Second):
		pids, _ := getChildPids(pgid)
		for _, pid := range pids {
			_ = killFunc(pid, syscall.SIGKILL)
		}
		_ = killFunc(-pgid, syscall.SIGKILL)
		return fmt.Errorf("process did not exit gracefully, killed")
	case err := <-done:
		return err
	}
}

func (r *ProcessRunner) Wait() error {
	if r.cmd == nil {
		return fmt.Errorf("process not started")
	}
	return r.cmd.Wait()
}

func (r *ProcessRunner) ExitCode() int {
	if r.cmd == nil || r.cmd.ProcessState == nil {
		return -1
	}
	return r.cmd.ProcessState.ExitCode()
}

func (r *ProcessRunner) Running() bool {
	return r.cmd != nil && r.cmd.Process != nil && r.cmd.ProcessState == nil
}
