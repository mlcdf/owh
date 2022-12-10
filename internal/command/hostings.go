package command

import (
	"strings"
)

type HostingsCommand struct {
	App
}

func (c *HostingsCommand) Help() string {
	helpText := `
Usage: owh hostings [options]

  List all your hostings
`
	return strings.TrimSpace(helpText)
}

func (c *HostingsCommand) Synopsis() string {
	return "List all your hostings"
}

func (c *HostingsCommand) Run(args []string) int {
	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	hostings, err := client.Hostings()
	if err != nil {
		return c.View.PrintErr(err)
	}

	tables := make([][]string, 0)

	for _, hosting := range hostings {
		row := []string{
			hosting.ServiceName,
			hosting.DisplayName,
			hosting.State,
			hosting.HostingIP,
			hosting.HostingIPv6,
			hosting.QuotaUsed.String(),
			hosting.QuotaSize.String(),
		}
		tables = append(tables, row)
	}

	err = c.View.Table("", tables, "Name", "Display Name", "State", "IPv4", "IPv6", "Disk used", "Disk available")
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
