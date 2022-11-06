package flow

import (
	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
)

func LinkFolder(client *api.Client, link *config.Link) error {
	hostings, err := client.ListHostings()
	if err != nil {
		return err
	}

	var selectedHosting string
	prompt := &survey.Select{Message: "Select a hosting", Options: hostings}
	err = survey.AskOne(prompt, &selectedHosting)
	if err != nil {
		return err
	}

	domains, err := client.ListDomains(selectedHosting)
	if err != nil {
		return err
	}

	domains = append(domains, "attach a new domain")

	var selectedDomain string
	prompt = &survey.Select{Message: "Select a domain", Options: domains}
	err = survey.AskOne(prompt, &selectedDomain)
	if err != nil {
		return err
	}

	if selectedDomain == "attach a new domain" {
		selectedDomain, err = AttachDomain(client, selectedHosting, "", false)
		if err != nil {
			return err
		}
	}

	link.Hosting = selectedHosting
	link.CanonicalDomain = selectedDomain

	// add local file to .gitignore
	return config.Save(link)
}
