package cmdutil

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
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

// Table renders the table defined by the given properties into w. Both title &
// cols are optional.
func Table(title string, rows [][]string, cols ...string) error {
	if title != "" {
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(title))
	}

	table := tablewriter.NewWriter(os.Stdout)

	if len(cols) > 0 {
		table.SetHeader(cols)
	}

	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnSeparator(" ")
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("\t")

	// table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})
	table.AppendBulk(rows)

	table.Render()

	fmt.Println()

	return nil
}
