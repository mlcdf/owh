package flow

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/view"
	"golang.org/x/xerrors"
)

func ListTasks(client *api.Client, view *view.View, hosting string) error {
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

	return view.Table("", tables, "ID", "Function", "Status", "Start date", "Last update")
}

func WaitTaskDone(client *api.Client, view *view.View, hosting string, id int64, message string) error {
	var task *api.Task
	t := time.Now()

	view.StartSpinner(message)
	defer view.StopSpinner()

	for {
		url := fmt.Sprintf("/hosting/web/%s/tasks/%d", hosting, id)
		err := client.Get(url, &task)
		if err != nil {
			var e *ovh.APIError
			if errors.As(err, &e) {
				if e.Code == http.StatusNotFound {
					// We arrive here when the task have been archived
					return nil
				}
			}
			return xerrors.Errorf("error fetching task status (task_id: %d): %w", id, err)
		}

		if task.Status == "done" {
			return nil
		}

		if task.Status == "error" || task.Status == "cancelled" {
			view.StopSpinner()
			view.Printf("Unexpected task status %s for ssh user creation (task_id: %d)\n", task.Status, id)
			return cmdutil.ErrSilent
		}

		if time.Since(t) > 5*time.Minute {
			view.StopSpinner()
			view.Printf("Timed out waiting (3 minutes) for %s task completion (task_id: %d)\n", task.Function, id)
			return cmdutil.ErrSilent
		}
	}
}
