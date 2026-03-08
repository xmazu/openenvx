package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/xmazu/oexctl/internal/proxy"
)

const proxyPort = proxy.DefaultPort

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Local dev proxy with *.localhost URLs",
	Long: `Proxy for local development: https://myapp.localhost:1355.
No DNS config needed – *.localhost resolves to 127.0.0.1 automatically.`,
}

var proxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the proxy daemon",
	RunE:  runProxyStop,
}

var proxyRunCmd = &cobra.Command{
	Use:   "run [name] -- [command]",
	Short: "Run app and register route",
	Long:  `Run a command and register it at https://<name>.localhost:1355. Same port (1355), different domain per app. First run starts the daemon. Use -- to separate name from command.`,
	RunE:  runProxyRun,
	// Don't parse flags so -- and command args pass through
	DisableFlagParsing: true,
}

func init() {
	proxyCmd.AddCommand(proxyStopCmd)
	proxyCmd.AddCommand(proxyRunCmd)
}

func runProxyStop(cmd *cobra.Command, args []string) error {
	state, err := proxy.NewState(proxyPort)
	if err != nil {
		return err
	}
	pid, err := state.ReadProxyPID()
	if err != nil {
		// No PID file (e.g. proxy started in-process by proxy run) - kill whatever is on the port
		killed, _ := proxy.KillProcessOnPort(proxyPort)
		if killed {
			fmt.Println("proxy stopped")
			return nil
		}
		return fmt.Errorf("proxy not running (no PID file, nothing on port %d)", proxyPort)
	}
	if err := proxy.KillProcess(pid); err != nil {
		return fmt.Errorf("kill proxy: %w", err)
	}
	state.RemoveProxyPID()
	fmt.Println("proxy stopped")
	return nil
}

func runProxyRun(cmd *cobra.Command, args []string) error {
	var name string
	var command []string
	var force bool
	beforeDash := []string{}
	dashIdx := -1
	for i, a := range args {
		if a == "--" {
			dashIdx = i
			if i+1 < len(args) {
				command = args[i+1:]
			}
			break
		}
		beforeDash = append(beforeDash, a)
	}
	if dashIdx < 0 {
		return fmt.Errorf("-- required to separate name from command (use: oexctl proxy run myapp -- npm run start)")
	}
	// beforeDash: [name] or [name, --force] or [--force, name]
	for _, a := range beforeDash {
		if a == "--force" {
			force = true
		} else if name == "" {
			name = a
		}
	}
	if name == "" {
		return fmt.Errorf("name required before -- (use: oexctl proxy run myapp -- npm run start)")
	}
	if len(command) == 0 {
		return fmt.Errorf("command required after -- (use: oexctl proxy run myapp -- npm run start)")
	}

	// Ensure proxy is running on 1355 (don't kill - port in use = assume our proxy)
	if !isProxyPortInUse() {
		srv, err := proxy.NewServerTLS(proxyPort)
		if err != nil {
			return fmt.Errorf("proxy: %w", err)
		}
		stateDir, _ := proxy.StateDirPath(proxyPort)
		_ = proxy.TrustCA(stateDir)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				fmt.Fprintf(os.Stderr, "proxy: %v\n", err)
			}
		}()
		for range 20 {
			time.Sleep(50 * time.Millisecond)
			if isProxyPortInUse() {
				break
			}
		}
		if !isProxyPortInUse() {
			return fmt.Errorf("proxy: failed to start")
		}
	}

	workdir, _ := os.Getwd()
	_, execCmd, err := proxy.Run(name, command, workdir, force)
	if err != nil {
		var conflict *proxy.RouteConflictError
		if errors.As(err, &conflict) {
			return fmt.Errorf("%w (use --force to override)", err)
		}
		return err
	}

	// On exit (signal or process), remove route
	defer func() {
		proxy.RemoveRoute(name)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		_ = proxy.KillProcess(execCmd.Process.Pid)
	}()

	err = execCmd.Wait()
	if err != nil && (errors.Is(err, os.ErrProcessDone) || isSignalExit(err)) {
		fmt.Println("gracefully shutting down")
		return nil
	}
	return err
}

func isSignalExit(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	return status.Signaled()
}

func isProxyPortInUse() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", proxyPort))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
