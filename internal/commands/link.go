package commands

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

func Link(client *api.Client, hosting string, domain string) error {
	link, err := config.NewLink()
	if err != nil && err != config.ErrFolderNotLinked {
		return err
	}

	if err != config.ErrFolderNotLinked {
		if hosting != "" && domain != "" && link.Hosting == hosting && link.CanonicalDomain == domain {
			fmt.Println("Nothing to do: already linked.")
			return nil
		}

		var shouldContinue bool
		err = survey.AskOne(&survey.Confirm{Message: "Current directory already linked. Do you want to continue?"}, &shouldContinue)
		if err != nil {
			return err
		}

		if !shouldContinue {
			return cmdutil.ErrCancel
		}
	}

	if hosting == "" {
		hosting, err = flow.SelectHosting(client, domain)
		if err != nil {
			return err
		}
	}

	if domain == "" {
		domain, err = flow.SelectDomain(client, hosting)
		if err != nil {
			return err
		}
	}

	link.Hosting = hosting
	link.CanonicalDomain = domain

	err = link.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Current directory linked to hosting %s and domain %s on OVHcloud\n", cmdutil.Highlight(hosting), cmdutil.Highlight(domain))
	return nil
}
