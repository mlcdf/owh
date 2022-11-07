package commands

import (
	"bytes"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"
)

func Remove(client *api.Client, hosting string, domain string) error {
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
		var shouldContinue bool
		prompt := &survey.Confirm{Message: fmt.Sprintf("Are you sure you want to remove %s on hosting %s", domain, hosting)}

		err := survey.AskOne(prompt, &shouldContinue)
		if err != nil {
			return err
		}

		if !shouldContinue {
			return cmdutil.ErrCancel
		}

		return nuke(client, hosting, &selectedDomain)
	}

	mightBeRelated = slices.Insert(mightBeRelated, 0, selectedDomain.Domain)
	mapDomains[selectedDomain.Domain] = selectedDomain

	prompt := &survey.MultiSelect{
		Message: "We found some domains that might be related. Select the ones to remove.",
		Options: mightBeRelated,
		Default: []string{selectedDomain.Domain},
	}
	var selectedDomains []string

	err = survey.AskOne(prompt, &selectedDomains)
	if err != nil {
		return err
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

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	err = session.Run(fmt.Sprintf("rm -rf %s", domain.Path))
	if err != nil {
		return xerrors.Errorf("failed to run command: %s : %w", b.String(), err)
	}

	err = client.DeleteDomain(hosting, domain.Domain)
	if err != nil {
		return err
	}

	return nil
}
