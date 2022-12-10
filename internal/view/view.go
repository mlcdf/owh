package view

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
)

var StyleSubtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var StyleHighlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
var StyleSpecial = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

type View struct {
	Writer io.Writer

	isInteractive bool
	spinner       *spinner.Spinner
}

type LabelValue struct {
	Label string
	Value string
}

func New(writer io.Writer, isInteractive bool) *View {
	view := &View{
		Writer:        writer,
		isInteractive: isInteractive,
	}

	if isInteractive {
		view.spinner = spinner.New(characterSet, 100*time.Millisecond)
	}

	return view
}

func (view *View) Println(a ...any) {
	fmt.Fprintln(view.Writer, a...)
}

func (view *View) Printf(format string, a ...any) {
	fmt.Fprintf(view.Writer, format, a...)
}

func (view *View) PrintErr(err error) int {
	fmt.Fprintln(view.Writer, err)
	return 1
}

// Table renders the table defined by the given properties into w. Both title &
// cols are optional.
func (view *View) Table(title string, rows [][]string, cols ...string) error {
	str := &strings.Builder{}

	table := tablewriter.NewWriter(str)

	if title != "" {
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(title))
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

	if title != "" {
		scanner := bufio.NewScanner(strings.NewReader(str.String()))

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(view.Writer, "  %s\n", line)
		}
	} else {
		fmt.Fprint(view.Writer, str)
	}

	return nil
}

func (view *View) VerticalTable(title string, rows []LabelValue) {
	var leftColumnSize int

	for _, row := range rows {
		if length := len(row.Label); leftColumnSize < length {
			leftColumnSize = length
		}
	}

	leftColumn := lipgloss.NewStyle().Width(leftColumnSize + 2).PaddingLeft(2)

	if title != "" {
		fmt.Fprintln(view.Writer, lipgloss.NewStyle().Bold(true).Render(title))
	}

	for _, row := range rows {
		fmt.Fprintf(view.Writer, "%s = %s\n", leftColumn.Render(row.Label), row.Value)
	}
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
