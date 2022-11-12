package flow

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
)

func SelectHosting(client *api.Client, domain string) (string, error) {
	if domain != "" {
		hosting, err := client.HostingByDomain(domain)
		if err == nil {
			return hosting, nil
		}

		if errors.Is(err, api.ErrNotHostingFound) {
			fmt.Printf("No associated hosting found for the domain %s\n", domain)
			return "", cmdutil.ErrSilent
		}

		if errors.Is(err, api.ErrMoreThanOneHostingFound) {
			return "", err
		}

		// we arrive here if err == MoreThanOneHostingFound
		fmt.Printf("More than one hosting found for the domain %s\n", domain)
	}

	hostings, err := client.ListHostings()
	if err != nil {
		return "", err
	}

	if len(hostings) == 1 {
		return hostings[0], nil
	}

	var selectedHosting string
	prompt := &survey.Select{Message: "Select a hosting", Options: hostings}
	err = survey.AskOne(prompt, &selectedHosting)
	if err != nil {
		return "", err
	}

	return selectedHosting, nil
}

func SelectDomain(client *api.Client, hosting string) (string, error) {
	domains, err := client.ListDomains(hosting)
	if err != nil {
		return "", err
	}

	var selectedDomain string
	prompt := &survey.Select{Message: "Select a domain", Options: domains}
	err = survey.AskOne(prompt, &selectedDomain)
	if err != nil {
		return "", err
	}

	return selectedDomain, nil
}

func LinkDirectory(client *api.Client, link *config.Link) error {
	selectedHosting, err := SelectHosting(client, "")
	if err != nil {
		return err
	}

	domains, err := client.ListDomains(selectedHosting)
	if err != nil {
		return err
	}

	domains = append(domains, "or attach a new domain")

	var selectedDomain string
	prompt := &survey.Select{Message: "Select a domain", Options: domains}
	err = survey.AskOne(prompt, &selectedDomain)
	if err != nil {
		return err
	}

	if selectedDomain == "or attach a new domain" {
		selectedDomain, err = AttachDomain(client, selectedHosting, "", false)
		if err != nil {
			return err
		}
	}

	link.Hosting = selectedHosting
	link.CanonicalDomain = selectedDomain

	// add local file to .gitignore
	return link.Save()
}
