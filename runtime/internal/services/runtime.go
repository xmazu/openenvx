package services

import "context"

// Status represents a service's current state
type Status struct {
	Name    string
	State   string // running, exited, etc.
	Ports   []string
	Healthy bool
}

// Runtime abstracts container operations
type Runtime interface {
	Start(ctx context.Context, composeFile string) error
	Stop(ctx context.Context, composeFile string) error
	Status(ctx context.Context, composeFile string) ([]Status, error)
	IsAvailable() bool
	Name() string
}
