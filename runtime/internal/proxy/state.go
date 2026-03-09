package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultPort = 1355
	StateDir    = "openenvx"

	lockStaleDuration = 10 * time.Second
	lockRetryCount    = 20
	lockRetryInterval = 50 * time.Millisecond
)

var ErrRouteLockFailed = errors.New("failed to acquire route lock")

type RouteConflictError struct {
	Name string
	PID  int
}

func (e *RouteConflictError) Error() string {
	return fmt.Sprintf("route %q already registered by another process (PID %d)", e.Name, e.PID)
}

type Route struct {
	Port int `json:"port"`
	PID  int `json:"pid"`
}

type State struct {
	mu     sync.RWMutex
	dir    string
	routes map[string]Route
}

// StateDirPath returns ~/.config/openenvx/
func StateDirPath(proxyPort int) (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, StateDir), nil
}

func NewState(proxyPort int) (*State, error) {
	dir, err := StateDirPath(proxyPort)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	s := &State{dir: dir, routes: make(map[string]Route)}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *State) lockPath() string {
	return filepath.Join(s.dir, "routes.lock")
}

func (s *State) acquireLock() error {
	lockPath := s.lockPath()
	for i := 0; i < lockRetryCount; i++ {
		err := os.Mkdir(lockPath, 0700)
		if err == nil {
			return nil
		}
		if !os.IsExist(err) {
			return fmt.Errorf("create lock: %w", err)
		}
		// Lock exists, check if stale
		info, err := os.Stat(lockPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("stat lock: %w", err)
		}
		if time.Since(info.ModTime()) > lockStaleDuration {
			_ = os.RemoveAll(lockPath)
			continue
		}
		time.Sleep(lockRetryInterval)
	}
	return ErrRouteLockFailed
}

func (s *State) releaseLock() {
	_ = os.RemoveAll(s.lockPath())
}

func (s *State) routesPath() string {
	return filepath.Join(s.dir, "routes.json")
}

func (s *State) pidPath() string {
	return filepath.Join(s.dir, "proxy.pid")
}

func isPidAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func (s *State) load() error {
	data, err := os.ReadFile(s.routesPath())
	if err != nil {
		return err
	}
	var raw map[string]Route
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes = make(map[string]Route)
	for name, r := range raw {
		if r.PID > 0 && isPidAlive(r.PID) {
			s.routes[name] = r
		}
	}
	return nil
}

func (s *State) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.routes, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	tmp := s.routesPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.routesPath()); err != nil {
		return err
	}
	return fixOwnership(s.dir, s.routesPath())
}

func (s *State) AddRoute(name string, port, pid int, force bool) error {
	if err := s.acquireLock(); err != nil {
		return err
	}
	defer s.releaseLock()

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return err
	}

	s.mu.Lock()
	if existing, ok := s.routes[name]; ok && isPidAlive(existing.PID) && !force {
		s.mu.Unlock()
		return &RouteConflictError{Name: name, PID: existing.PID}
	}
	s.routes[name] = Route{Port: port, PID: pid}
	s.mu.Unlock()
	return s.save()
}

func (s *State) RemoveRoute(name string) (Route, bool) {
	if err := s.acquireLock(); err != nil {
		return Route{}, false
	}
	defer s.releaseLock()

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return Route{}, false
	}

	s.mu.Lock()
	r, ok := s.routes[name]
	if ok {
		delete(s.routes, name)
	}
	s.mu.Unlock()
	if ok {
		_ = s.save()
	}
	return r, ok
}

func (s *State) Reload() error {
	return s.load()
}

func (s *State) GetRoute(name string) (Route, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.routes[name]
	return r, ok
}

func (s *State) AllRoutes() map[string]Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Route, len(s.routes))
	for k, v := range s.routes {
		out[k] = v
	}
	return out
}

func (s *State) WriteProxyPID(pid int) error {
	path := s.pidPath()
	if err := os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return err
	}
	return fixOwnership(s.dir, path)
}

func (s *State) ReadProxyPID() (int, error) {
	data, err := os.ReadFile(s.pidPath())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

func (s *State) RemoveProxyPID() error {
	return os.Remove(s.pidPath())
}
