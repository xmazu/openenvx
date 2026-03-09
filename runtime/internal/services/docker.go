package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type Docker struct{}

func NewDocker() *Docker {
	return &Docker{}
}

func (d *Docker) Name() string {
	return "docker"
}

func (d *Docker) IsAvailable() bool {
	cmd := exec.Command("docker", "compose", "version")
	return cmd.Run() == nil
}

func (d *Docker) Start(ctx context.Context, composeFile string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "up", "-d")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w\n%s", err, out)
	}
	return nil
}

func (d *Docker) Stop(ctx context.Context, composeFile string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "down")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w\n%s", err, out)
	}
	return nil
}

type container struct {
	Service    string `json:"Service"`
	State      string `json:"State"`
	Publishers []struct {
		TargetPort    int `json:"TargetPort"`
		PublishedPort int `json:"PublishedPort"`
	} `json:"Publishers"`
}

func (d *Docker) Status(ctx context.Context, composeFile string) ([]Status, error) {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "ps", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker compose ps failed: %w", err)
	}

	var containers []container
	if err := json.Unmarshal(out, &containers); err != nil {
		var single container
		if err := json.Unmarshal(out, &single); err != nil {
			return []Status{}, nil
		}
		containers = []container{single}
	}

	result := make([]Status, 0, len(containers))
	for _, c := range containers {
		ports := make([]string, 0, len(c.Publishers))
		for _, p := range c.Publishers {
			if p.PublishedPort > 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", p.PublishedPort, p.TargetPort))
			}
		}
		result = append(result, Status{
			Name:    c.Service,
			State:   c.State,
			Ports:   ports,
			Healthy: c.State == "running",
		})
	}
	return result, nil
}
