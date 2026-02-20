package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func PlaintextInput(title string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(title).
		Value(&result).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}
	return result, nil
}
