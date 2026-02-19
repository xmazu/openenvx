package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

type MockPrompts struct {
	ConfirmFunc      func(title string) (bool, error)
	SelectStringFunc func(title string, options []StringOption) (string, error)
	InputFunc        func(title string) (string, error)
	HiddenInputFunc  func(title string) (string, error)
}

var activeMock *MockPrompts

type StringOption struct {
	Key   string
	Value string
}

func HiddenInput(title string) (string, error) {
	if activeMock != nil && activeMock.HiddenInputFunc != nil {
		return activeMock.HiddenInputFunc(title)
	}
	var result string
	err := huh.NewInput().
		Title(title).
		EchoMode(huh.EchoModePassword).
		Value(&result).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}
	return result, nil
}
