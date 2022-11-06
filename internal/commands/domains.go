package commands

import (
	"fmt"
	"strings"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

func ListDomains(client *api.Client, hosting string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			if err != config.ErrFolderNotLinked {
				return err
			}
			fmt.Println("Folder not link. Please run: owh link first")
		}
		hosting = link.Hosting
	}

	// TODO: print warning for domain where Path does not exists anymore

	domains, err := client.ListDomains(hosting)
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(domains, "\n"))
	return nil
}

func AttachDomain(client *api.Client, hosting string, domain string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			if err != config.ErrFolderNotLinked {
				return err
			}
			fmt.Println("Folder not link. Please run: owh link first")
		}

		hosting = link.Hosting
	}

	_, err := flow.AttachDomain(client, hosting, domain, false)
	if err != nil {
		return err
	}

	return nil
}

func DetachDomain(client *api.Client, hosting string, domain string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			if err != config.ErrFolderNotLinked {
				return err
			}
			fmt.Println("Folder not link. Please run: owh link first")
		}

		hosting = link.Hosting
	}

	err := flow.DetachDomain(client, hosting, domain)
	if err != nil {
		return err
	}

	return nil
}
