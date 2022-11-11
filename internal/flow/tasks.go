package flow

import (
	"fmt"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
)

func ListTasks(client *api.Client, hosting string) (string, error) {
	tasks, err := client.Tasks(hosting)
	if err != nil {
		return "", err
	}

	if len(tasks) == 0 {
		return "", nil
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
