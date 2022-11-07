package commands

import (
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
)

func Hosting(client *api.Client, hosting string) error {
	hostings, err := client.Hostings()
	if err != nil {
		return err
	}

	tables := make([][]string, 0)

	for _, hosting := range hostings {
		row := []string{
			hosting.ServiceName,
			hosting.DisplayName,
			hosting.State,
			hosting.HostingIp,
			hosting.HostingIpv6,
			hosting.QuotaUsed.String(),
			hosting.QuotaSize.String(),
		}
		tables = append(tables, row)
	}

	return cmdutil.Table("", tables, "Name", "Display Name", "State", "IPv4", "IPv6", "Disk used", "Disk available")
}
