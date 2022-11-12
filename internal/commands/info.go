package commands

import (
	"fmt"
	"strconv"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

func Info(client *api.Client) error {
	link, err := config.EnsureLink()
	if err != nil {
		return err
	}

	hosting, err := client.GetHosting(link.Hosting)
	if err != nil {
		return err
	}

	domain, err := client.GetDomain(link.Hosting, link.CanonicalDomain)
	if err != nil {
		return err
	}

	domains, err := flow.DomainsWithPath(client, hosting.ServiceName, domain)
	if err != nil {
		return err
	}

	hostingInfo := []cmdutil.LabelValue{
		{Label: "Service name", Value: hosting.ServiceName},
		{Label: "Display name", Value: hosting.DisplayName},
		{Label: "IPv4", Value: hosting.HostingIP},
		{Label: "IPv6", Value: hosting.HostingIPv6},
		{Label: "Disk used", Value: hosting.QuotaUsed.String()},
		{Label: "Disk available", Value: hosting.QuotaSize.String()},
	}

	cmdutil.DescriptionTable("Web hosting", hostingInfo)

	fmt.Println()

	tables := make([][]string, 0)

	for _, domain := range domains {
		name := domain.Domain

		if domain.Domain == link.CanonicalDomain {
			name = cmdutil.Special(domain.Domain)
		}

		row := []string{
			name,
			domain.Path,
			strconv.FormatBool(domain.SSL),
			domain.Firewall,
		}
		tables = append(tables, row)
	}

	err = cmdutil.PrintTable("Domains", tables, "Name", "Path", "SSL", "Firewall")
	if err != nil {
		return err
	}

	fmt.Println()

	users, err := client.ListUsers(link.Hosting)
	if err != nil {
		return err
	}

	tables = nil
	for _, user := range users {
		var primaryLogin bool

		if hosting.PrimaryLogin == user {
			primaryLogin = true
		}

		if credentials, ok := config.GlobalOpts.SFTPCredentials[link.Hosting]; ok && credentials.User == user {
			user = cmdutil.Special(user)
		}

		row := []string{
			user,
			strconv.FormatBool(primaryLogin),
		}
		tables = append(tables, row)
	}

	err = cmdutil.PrintTable("Users", tables, "Login", "Primary login")
	if err != nil {
		return err
	}

	fmt.Println()

	tasks, err := client.Tasks(link.Hosting)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return nil
	}

	tables = nil

	for _, task := range tasks {
		row := []string{
			fmt.Sprintf("%d", task.ID),
			task.Function,
			task.Status,
			task.StartDate.String(),
			task.LastUpdate.String(),
		}
		tables = append(tables, row)
	}

	err = cmdutil.PrintTable("Tasks", tables, "ID", "Function", "Status", "Start date", "Last update")
	if err != nil {
		return err
	}

	return nil
}
