package command

import (
	"fmt"
	"strconv"
	"strings"

	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/flow"
	"go.mlcdf.fr/owh/internal/unit"
	"go.mlcdf.fr/owh/internal/view"
)

type InfoCommand struct {
	App
}

func (c *InfoCommand) Help() string {
	helpText := `
Usage: owh info

  Show info about the linked website
`
	return strings.TrimSpace(helpText)
}

func (c *InfoCommand) Synopsis() string {
	return "Show info about the linked website"
}

func (c *InfoCommand) Run(args []string) int {
	link, err := c.EnsureLink()
	if err != nil {
		return c.View.PrintErr(err)
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	hosting, err := client.GetHosting(link.Hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	domain, err := client.GetDomain(link.Hosting, link.CanonicalDomain)
	if err != nil {
		return c.View.PrintErr(err)
	}

	domains, err := flow.DomainsWithPath(client, hosting.ServiceName, domain)
	if err != nil {
		return c.View.PrintErr(err)
	}

	hostingInfo := []view.LabelValue{
		{Label: "Service name", Value: hosting.ServiceName},
		{Label: "Display name", Value: hosting.DisplayName},
		{Label: "IPv4", Value: hosting.HostingIP},
		{Label: "IPv6", Value: hosting.HostingIPv6},
		{Label: "Disk quota", Value: hosting.QuotaUsed.String() + " / " + hosting.QuotaSize.String()},
	}

	quota, err := unit.Quota(hosting.QuotaUsed, hosting.QuotaSize)
	if err == nil && quota > 0.01 {
		hostingInfo[len(hostingInfo)-1].Value += fmt.Sprintf(" (%.2f)", quota)
	}

	c.View.VerticalTable("Web hosting", hostingInfo)

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

	err = c.View.Table("Domains", tables, "Name", "Path", "SSL", "Firewall")
	if err != nil {
		return c.View.PrintErr(err)
	}
	fmt.Println()

	users, err := client.ListUsers(link.Hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}
	tables = nil
	for _, user := range users {
		var primaryLogin bool

		if hosting.PrimaryLogin == user {
			primaryLogin = true
		}

		if credentials, ok := c.Config.SFTPCredentials[link.Hosting]; ok && credentials.User == user {
			user = cmdutil.Special(user)
		}

		row := []string{
			user,
			strconv.FormatBool(primaryLogin),
		}
		tables = append(tables, row)
	}

	err = c.View.Table("Users", tables, "Login", "Primary login")
	if err != nil {
		return c.View.PrintErr(err)
	}

	fmt.Println()

	tasks, err := client.Tasks(link.Hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	if len(tasks) == 0 {
		return 0
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

	err = c.View.Table("Tasks", tables, "ID", "Function", "Status", "Start date", "Last update")
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
