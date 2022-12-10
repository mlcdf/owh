package api

import (
	"context"
	"fmt"
	"time"

	"github.com/alitto/pond"
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
