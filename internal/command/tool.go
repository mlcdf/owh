package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type ToolCommand struct {
	App
}

func (c *ToolCommand) Help() string {
	helpText := `
Usage: owh tool [--help] <command> [<args>]
	
  This command groups useful extra subcommands.
`
	return strings.TrimSpace(helpText)
}

func (c *ToolCommand) Synopsis() string {
	return "Group useful extra-commands"
}

func (c *ToolCommand) Run(args []string) int {
	return cli.RunResultHelp
}
