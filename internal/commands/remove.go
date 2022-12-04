package commands

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"
)

func Remove(client *api.Client, hosting string, domain string, yes bool) error {
	if hosting == "" && domain == "" {
		link, err := config.EnsureLink()
		if err != nil {
			return err
		}

		hosting = link.Hosting
		domain = link.CanonicalDomain
	}

	domains, err := client.Domains(hosting)
	if err != nil {
		return err
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
		if cmdutil.IsInteractive() && !yes {
			var shouldContinue bool

			prompt := &survey.Confirm{Message: fmt.Sprintf("Are you sure you want to remove %s on hosting %s", domain, hosting)}

			err := survey.AskOne(prompt, &shouldContinue)
			if err != nil {
				return err
			}

			if !shouldContinue {
				return cmdutil.ErrCancel
			}
		}

		return nuke(client, hosting, &selectedDomain)
	}

	mightBeRelated = slices.Insert(mightBeRelated, 0, selectedDomain.Domain)
	mapDomains[selectedDomain.Domain] = selectedDomain

	var selectedDomains []string

	if cmdutil.IsInteractive() && !yes {
		prompt := &survey.MultiSelect{
			Message: "We found some domains that might be related. Select the ones to remove.",
			Options: mightBeRelated,
			Default: []string{selectedDomain.Domain},
		}

		err = survey.AskOne(prompt, &selectedDomains)
		if err != nil {
			return err
		}
	} else {
		selectedDomains = mightBeRelated
	}

	for _, domain := range selectedDomains {
		d := mapDomains[domain]

		err := nuke(client, hosting, &d)
		if err != nil {
			return err
		}
		fmt.Printf("%s removed\n", domain)
	}

	return nil
}

func nuke(client *api.Client, hosting string, domain *api.AttachedDomain) error {
	conn, err := flow.NewSSHClient(client, hosting)
	if err != nil {
		return err
	}

	err = conn.ForceRemove(domain.Path)
	if err != nil {
		return xerrors.Errorf("failed remove %s : %w", domain.Path, err)
	}

	err = client.DeleteDomain(hosting, domain.Domain)
	if err != nil {
		return err
	}

	return nil
}
