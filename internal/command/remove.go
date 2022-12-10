package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
	"go.mlcdf.fr/owh/internal/view"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"
)

type RemoveCommand struct {
	App
}

func (c *RemoveCommand) Help() string {
	helpText := `
Usage: owh remove, rm [<options>]

  Removes websites (files & attached domains).

Options:
  --hosting       service name (if not set, you'll be prompt)
  --domain        domain name (if not set, you'll be prompt)
  --yes
`
	return strings.TrimSpace(helpText)
}

func (c *RemoveCommand) Synopsis() string {
	return "Remove websites (files & attached domains)"
}

func (c *RemoveCommand) Run(args []string) int {
	var hosting string
	var domain string
	var yes bool

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")
	flags.StringVar(&domain, "domain", "", "")
	flags.BoolVar(&yes, "yes", false, "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	if hosting == "" && domain == "" {
		link, err := c.EnsureLink()
		if err != nil {
			return c.View.PrintErr(err)
		}

		hosting = link.Hosting
		domain = link.CanonicalDomain
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	domains, err := client.Domains(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	var selectedDomain api.AttachedDomain

	for _, d := range domains {
		if d.Domain == domain {
			selectedDomain = d
		}
	}

	mapDomains := make(map[string]api.AttachedDomain, 0)
	mightBeRelated := make([]string, 0)

	for _, d := range domains {
		if selectedDomain.Domain != d.Domain && selectedDomain.Path == d.Path {
			mapDomains[d.Domain] = d
			mightBeRelated = append(mightBeRelated, d.Domain)
		}
	}

	if len(mightBeRelated) == 0 {
		if c.IsInteractive && !yes {
			var shouldContinue bool

			prompt := &survey.Confirm{Message: fmt.Sprintf("Are you sure you want to remove %s on hosting %s", domain, hosting)}

			err := survey.AskOne(prompt, &shouldContinue)
			if err != nil {
				c.View.PrintErr(err)
			}

			if !shouldContinue {
				return 2
			}
		}

		err = nuke(client, c.Config, c.View, c.IsInteractive, hosting, &selectedDomain)
		if err != nil {
			c.View.PrintErr(err)
		}

		return 0
	}

	mightBeRelated = slices.Insert(mightBeRelated, 0, selectedDomain.Domain)
	mapDomains[selectedDomain.Domain] = selectedDomain

	var selectedDomains []string

	if c.IsInteractive && !yes {
		prompt := &survey.MultiSelect{
			Message: "We found some domains that might be related. Select the ones to remove.",
			Options: mightBeRelated,
			Default: []string{selectedDomain.Domain},
		}

		err = survey.AskOne(prompt, &selectedDomains)
		if err != nil {
			c.View.PrintErr(err)
		}
	} else {
		selectedDomains = mightBeRelated
	}

	for _, domain := range selectedDomains {
		d := mapDomains[domain]

		err := nuke(client, c.Config, c.View, c.IsInteractive, hosting, &d)
		if err != nil {
			c.View.PrintErr(err)
		}
		fmt.Printf("%s removed\n", domain)
	}

	return 0
}

func nuke(client *api.Client, config *config.Config, view *view.View, isInteractive bool, hosting string, domain *api.AttachedDomain) error {
	conn, err := flow.NewSSHClient(client, config, view, isInteractive, hosting)
	if err != nil {
		return err
	}

	err = conn.ForceRemove(domain.Path)
	if err != nil {
		return xerrors.Errorf("failed remove %s : %w", domain.Path, err)
	}

	_, err = client.DeleteDomain(hosting, domain.Domain)
	if err != nil {
		return err
	}

	return nil
}
