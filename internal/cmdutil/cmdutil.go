package cmdutil

import (
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/xerrors"
)

// ErrSilent is an error that triggers exit code 1 without any error messaging.
var ErrSilent = xerrors.New("")

// ErrCancel signals user-initiated cancellation.
var ErrCancel = xerrors.New("")

var StyleSubtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var StyleHighlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
var StyleSpecial = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

func Color(c lipgloss.TerminalColor) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

func Subtle(str string) string {
	return lipgloss.NewStyle().Foreground(StyleSubtle).Render(str)
}

func Highlight(str string) string {
	return lipgloss.NewStyle().Foreground(StyleHighlight).Render(str)
}

func Special(str string) string {
	return lipgloss.NewStyle().Foreground(StyleSpecial).Render(str)
}

func Bold(str string) string {
	return lipgloss.NewStyle().Bold(true).Render(str)
}
