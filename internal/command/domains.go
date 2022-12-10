package command

import (
	"flag"
	"strconv"
	"strings"
)

type DomainsCommand struct {
	App
}

func (c *DomainsCommand) Help() string {
	helpText := `
Usage: owh domains [<command>]

  Handles various domain operations.
  Lists all the domains when run without subcommand.
`
	return strings.TrimSpace(helpText)
}

func (c *DomainsCommand) Synopsis() string {
	return "Handle various domain operations"
}

func (c *DomainsCommand) Run(args []string) int {
	var hosting string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	if hosting == "" {
		link, err := c.EnsureLink()
		if err != nil {
			return c.View.PrintErr(err)
		}

		hosting = link.Hosting
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	domains, err := client.Domains(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	if len(domains) == 0 {
		return 0
	}

	tables := make([][]string, 0)

	for _, domain := range domains {
		row := []string{
			domain.Domain,
			domain.Path,
			strconv.FormatBool(domain.SSL),
			domain.Firewall,
		}
		tables = append(tables, row)
	}

	err = c.View.Table("", tables, "Domain", "Path", "SSL", "Firewall")
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
