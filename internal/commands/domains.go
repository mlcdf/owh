package commands

import (
	"fmt"
	"strconv"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

func ListDomains(client *api.Client, hosting string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			return err
		}

		hosting = link.Hosting
	}

	domains, err := client.Domains(hosting)
	if err != nil {
		return err
	}

	if len(domains) == 0 {
		return nil
	}

	tables := make([][]string, 0)

	for _, domain := range domains {
		row := []string{
			domain.Domain,
			domain.Path,
			strconv.FormatBool(domain.SSL),
			domain.Firewall,
		}
		tables = append(tables, row)
	}

	return cmdutil.PrintTable("", tables, "Domain", "Path", "SSL", "Firewall")
}

func AttachDomain(client *api.Client, hosting string, domain string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil && err != config.ErrFolderNotLinked {
			return err
		}

		hosting = link.Hosting
	}

	if domain == "" && !cmdutil.IsInteractive() {
		fmt.Println("missing flag --domain")
		return cmdutil.ErrFlag
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
		if err != nil && err != config.ErrFolderNotLinked {
			return err
		}

		hosting = link.Hosting
	}

	if domain == "" && !cmdutil.IsInteractive() {
		fmt.Println("missing flag --domain")
		return cmdutil.ErrFlag
	}

	err := flow.DetachDomain(client, hosting, domain)
	if err != nil {
		return err
	}

	return nil
}
