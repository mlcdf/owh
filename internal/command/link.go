package command

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

type LinkCommand struct {
	App
}

func (c *LinkCommand) Help() string {
	helpText := `
Usage: owh link

  Link the current directory to an existing website on OVHcloud.

  To link to a new website, use 'owh deploy'.

Options:

  --hosting       service name (if not set, you'll be prompt)
  --domain        domain name (if not set, you'll be prompt)
`
	return strings.TrimSpace(helpText)
}

func (c *LinkCommand) Synopsis() string {
	return "Link current directory to an existing website on OVHcloud"
}

func (c *LinkCommand) Run(args []string) int {
	var hosting string
	var domain string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")
	flags.StringVar(&domain, "domain", "", "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	link, err := c.EnsureLink()
	if err != nil && !errors.Is(err, config.ErrFolderNotLinked) {
		return c.View.PrintErr(err)
	}

	if err != config.ErrFolderNotLinked {
		if hosting != "" && domain != "" && link.Hosting == hosting && link.CanonicalDomain == domain {
			fmt.Println("Nothing to do: already linked.")
			return 0
		}

		var shouldContinue bool
		err = survey.AskOne(
			&survey.Confirm{Message: "Current directory already linked. Do you want to continue?"},
			&shouldContinue,
		)
		if err != nil {
			return c.View.PrintErr(err)
		}

		if !shouldContinue {
			return 2
		}
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	if hosting == "" {
		hosting, err = flow.SelectHosting(client, domain)
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	if domain == "" {
		domain, err = flow.SelectDomain(client, hosting)
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	link.Hosting = hosting
	link.CanonicalDomain = domain

	err = link.Save()
	if err != nil {
		return c.View.PrintErr(err)
	}

	fmt.Printf(
		"Linked to hosting %s and domain %s on OVHcloud (created .owh.json and added it to .gitignore)\n",
		cmdutil.Highlight(hosting),
		cmdutil.Highlight(domain),
	)
	return 0
}
