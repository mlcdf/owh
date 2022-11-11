package flow

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ovh/go-ovh/ovh"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
)

func AttachDomain(client *api.Client, hosting string, domain string, www bool) (string, error) {
	if domain == "" {
		prompt := &survey.Input{Message: "Enter a domain name"}
		err := survey.AskOne(prompt, &domain, survey.WithValidator(survey.Required))
		if err != nil {
			return "", err
		}
	}

	err := attachDomain(client, hosting, domain, domain)
	if err != nil {
		return "", err
	}

	if www {
		suggestedDomain := suggestDomain(domain)

		err = attachDomain(client, hosting, suggestedDomain, domain)
		if err != nil {
			return "", err
		}
		fmt.Printf("Domains %s and %s attached\n", cmdutil.Color(cmdutil.StyleHighlight).Render(domain), cmdutil.Color(cmdutil.StyleHighlight).Render(suggestedDomain))
	} else {
		fmt.Printf("Domain %s attached\n", cmdutil.Color(cmdutil.StyleHighlight).Render(domain))
	}

	return domain, nil
}

func suggestDomain(domain string) string {
	if strings.HasPrefix(domain, "www") {
		return strings.Replace(domain, "www.", "", 1)
	} else {
		return "www." + domain
	}
}

func attachDomain(client *api.Client, hosting string, domain string, path string) error {
	attachedDomain, err := client.GetDomain(hosting, domain)
	if err != nil {

		var e *ovh.APIError
		if errors.As(err, &e) && e.Code == 404 {
			err := client.PostDomain(hosting, domain)
			if err == nil {
				return nil
			}
		}

		return err
	}

	if attachedDomain.Path != path {
		err := client.UpdateDomain(hosting, domain)
		if err != nil {
			return err
		}
	}

	return nil
}

func DetachDomain(client *api.Client, hosting string, domain string) error {
	if domain == "" {
		prompt := &survey.Input{Message: "Enter a domain name"}
		err := survey.AskOne(prompt, &domain, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	suggestedDomain := suggestDomain(domain)
	var addSuggestedDomain bool

	prompt2 := &survey.Confirm{Message: fmt.Sprintf("Would you like to also detached domain %s", suggestedDomain)}
	err := survey.AskOne(prompt2, &addSuggestedDomain)
	if err != nil {
		return err
	}

	err = client.DeleteDomain(hosting, domain)
	if err != nil {
		return err
	}

	if addSuggestedDomain {
		err = client.DeleteDomain(hosting, suggestedDomain)
		if err != nil {
			return err
		}
		fmt.Printf("Domains %s and %s Detached\n", cmdutil.Color(cmdutil.StyleHighlight).Render(hosting), cmdutil.Color(cmdutil.StyleHighlight).Render(suggestedDomain))
	} else {
		fmt.Printf("Domain %s Detached\n", cmdutil.Color(cmdutil.StyleHighlight).Render(hosting))
	}

	return nil
}

func DomainsWithPath(client *api.Client, hosting string, domain *api.AttachedDomain) ([]api.AttachedDomain, error) {
	domains, err := client.Domains(hosting)
	if err != nil {
		return nil, err
	}

	out := []api.AttachedDomain{}

	for _, d := range domains {
		if d.Path == domain.Path {
			out = append(out, d)
		}
	}

	return out, nil
}
