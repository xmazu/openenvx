package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")).
			Bold(true)

	SeverityCritical = lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Bold(true)

	SeverityHigh = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	SeverityMedium = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	SeverityLow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("4"))
)

func Header(text string) string {
	return HeaderStyle.Render(text)
}

func Success(text string) string {
	return SuccessStyle.Render(text)
}

func Warning(text string) string {
	return WarningStyle.Render(text)
}

func Error(text string) string {
	return ErrorStyle.Render(text)
}

func Muted(text string) string {
	return MutedStyle.Render(text)
}

func Key(text string) string {
	return KeyStyle.Render(text)
}

func Label(text string) string {
	return LabelStyle.Render(text)
}

func FormatKeyDisplay(key string) string {
	if len(key) <= 20 {
		return Key(key)
	}
	return Key(key[:12] + "...")
}
