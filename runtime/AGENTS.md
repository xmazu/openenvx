# Runtime (oexctl) - Agent Guidelines

## Architecture Principles

### Command Layer (cmd/)
- **Thin commands** - Parse args, validate inputs, call internal packages
- **NO business logic** - Delegates to internal/
- **Error handling** - Wrap and present errors to user

### Internal Layer (internal/)
- **All business logic lives here**
- **proxy/** - Proxy routing and TLS (existing)
- **services/** - Container/service management (Docker Compose, future Podman)

## Example Pattern

```go
// cmd/up.go - thin wrapper
cmd := &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        mgr, err := services.NewManager(cwd)
        if err != nil {
            return err
        }
        return mgr.Start(ctx) // Delegate to internal/
    },
}

// internal/services/manager.go - all logic here
func (m *Manager) Start(ctx context.Context) error {
    // Docker compose logic, error handling, etc.
}
```

## Package Guidelines

**services/ responsibilities:**
- Runtime abstraction (Docker now, Podman later)
- Compose file management
- Container lifecycle (start, stop, status)
- PID/state tracking

**Commands to implement:**
- `oexctl up` - Start proxy + services
- `oexctl down` - Stop services (keep proxy running? or stop all?)
- Future: `oexctl services status` if needed

## State Management

Store runtime state in `.openenvx/`:
- `services.yaml` - Docker Compose definition
- `state.json` - Existing proxy state
- `pids.json` - Service process tracking (if needed)

## Testing

- Unit tests in `internal/services/*_test.go`
- Mock Runtime for CI
- Integration tests require Docker
