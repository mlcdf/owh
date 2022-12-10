package command

import (
	"flag"
	"fmt"
	"strings"

	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

type DetachCommand struct {
	App
}

func (c *DetachCommand) Help() string {
	helpText := `
Usage: owh domains [options]

  Deploys the linked website to OVHcloud Web Hosting.
  If the directory is not linked, it'll ask to linked it to a hosting first.

Options:

  --hosting       service name (if not set, you'll be prompt)
  --domain        domain name (if not set, you'll be prompt)
	`
	return strings.TrimSpace(helpText)
}

func (c *DetachCommand) Synopsis() string {
	return "Detach a domain"
}

func (c *DetachCommand) Run(args []string) int {
	var hosting string
	var domain string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")
	flags.StringVar(&domain, "domain", "", "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	if hosting == "" {
		link, err := c.EnsureLink()
		if err != nil && err != config.ErrFolderNotLinked {
			return c.View.PrintErr(err)
		}

		hosting = link.Hosting
	}

	if domain == "" && !c.IsInteractive {
		fmt.Println("missing flag --domain")
		return 1
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	err = flow.DetachDomain(client, hosting, domain)
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
