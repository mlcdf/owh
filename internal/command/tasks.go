package command

import (
	"flag"
	"fmt"
	"strings"
)

type TasksCommand struct {
	App
}

func (c *TasksCommand) Help() string {
	helpText := `
Usage: owh tasks

  Lists tasks.
`
	return strings.TrimSpace(helpText)
}

func (c *TasksCommand) Synopsis() string {
	return "List tasks"
}

func (c *TasksCommand) Run(args []string) int {
	var hosting string

	flags := flag.NewFlagSet("tasks", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")

	err := flags.Parse(args)
	if err != nil {
		return c.View.PrintErr(err)
	}

	if hosting == "" {
		link, err := c.EnsureLink()
		if err != nil {
			return c.View.PrintErr(err)
		}

		hosting = link.Hosting
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	tasks, err := client.Tasks(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	if len(tasks) == 0 {
		return 0
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

	err = c.View.Table("", tables, "ID", "Function", "Status", "Start date", "Last update")
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
