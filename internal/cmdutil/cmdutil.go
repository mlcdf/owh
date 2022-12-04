package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/xerrors"
)

// ErrSilent is an error that triggers exit code 1 without any error messaging.
var ErrSilent = xerrors.New("ErrSilent")

// ErrCancel signals user-initiated cancellation.
var ErrCancel = xerrors.New("ErrCancel")

// ErrFlag is an error of flags or arguments (missing argument, invalid flag, etc...).
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

func Bold(str string) string {
	return lipgloss.NewStyle().Bold(true).Render(str)
}

func PrintTable(title string, rows [][]string, cols ...string) error {
	str, err := Table(title, rows, cols...)
	if err != nil {
		return err
	}

	if title != "" {
		scanner := bufio.NewScanner(strings.NewReader(str))

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("  %s\n", line)
		}
	} else {
		fmt.Println(str)
	}

	return nil
}

// Table renders the table defined by the given properties into w. Both title &
// cols are optional.
func Table(title string, rows [][]string, cols ...string) (string, error) {
	tableString := &strings.Builder{}

	table := tablewriter.NewWriter(tableString)

	if title != "" {
		fmt.Println(Bold(title))
	}

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

	table.AppendBulk(rows)

	table.Render()

	return tableString.String(), nil
}

type LabelValue struct {
	Label string
	Value string
}

func DescriptionTable(title string, rows []LabelValue) {
	var leftColumnSize int

	for _, row := range rows {
		if length := len(row.Label); leftColumnSize < length {
			leftColumnSize = length
		}
	}

	leftColumn := lipgloss.NewStyle().Width(leftColumnSize + 2).PaddingLeft(2)

	if title != "" {
		fmt.Println(Bold(title))
	}

	for _, row := range rows {
		fmt.Printf("%s = %s\n", leftColumn.Render(row.Label), row.Value)
	}
}
