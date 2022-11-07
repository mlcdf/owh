package commands

import (
	"fmt"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
)

func Tasks(client *api.Client, hosting string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			return err
		}

		hosting = link.Hosting
	}

	tasks, err := client.Tasks(hosting)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return nil
	}

	tables := make([][]string, 0)

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

	return cmdutil.Table("", tables, "ID", "Function", "Status", "Start date", "Last update")
}
