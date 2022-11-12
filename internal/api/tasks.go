package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alitto/pond"
	"github.com/ovh/go-ovh/ovh"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/cmdutil/spinner"
	"golang.org/x/xerrors"
)

type Task struct {
	ID         int64     `json:"id"`
	Status     string    `json:"status"`
	Function   string    `json:"function"`
	StartDate  time.Time `json:"startDate"`
	LastUpdate time.Time `json:"lastUpdate"`
}

func (client *Client) ListTasks(hosting string) ([]int, error) {
	var taskIds []int

	url := fmt.Sprintf("/hosting/web/%s/tasks", hosting)

	err := client.Get(url, &taskIds)
	if err != nil {
		return nil, xerrors.Errorf("failed to GET %s: %w", url, err)
	}

	return taskIds, nil
}

func (client *Client) Tasks(hosting string) ([]*Task, error) {
	tasksIds, err := client.ListTasks(hosting)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	pool := pond.New(20, 20)
	defer pool.StopAndWait()

	// Create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	for _, id := range tasksIds {
		url := fmt.Sprintf("/hosting/web/%s/tasks/%d", hosting, id)

		group.Submit(func() error {
			var t *Task
			err := client.Get(url, &t)
			if err != nil {
				return err
			}

			tasks = append(tasks, t)

			return nil
		})
	}

	// Wait for all HTTP requests to complete.
	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (client *Client) WaitTaskDone(hosting string, id int64, message string) error {
	var task *Task
	t := time.Now()

	if cmdutil.IsInteractive() {
		spinner.Start(message)
		defer spinner.Stop()
	}

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
			spinner.Stop()
			fmt.Printf("Unexpected task status %s for ssh user creation (task_id: %d)\n", task.Status, id)
			return cmdutil.ErrSilent
		}

		if time.Since(t) > 5*time.Minute {
			spinner.Stop()
			fmt.Printf("Timed out waiting (3 minutes) for %s task completion (task_id: %d)\n", task.Function, id)
			return cmdutil.ErrSilent
		}
	}
}
