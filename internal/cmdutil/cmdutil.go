package cmdutil

import (
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/xerrors"

	"github.com/charmbracelet/lipgloss"
)

// ErrSilent is an error that triggers exit code 1 without any error messaging
var ErrSilent = xerrors.New("ErrSilent")

// ErrCancel signals user-initiated cancellation
var ErrCancel = xerrors.New("ErrCancel")

// ErrFlag is an error of flags or arguments (missing argument, invalid flag, etc...)
var ErrFlag = xerrors.New("ErrFlag")

var StyleSubtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var StyleHighlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
var StyleSpecial = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

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
