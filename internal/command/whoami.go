package command

import (
	"fmt"
	"strings"

	"go.mlcdf.fr/owh/internal/view"
)

type WhoamiCommand struct {
	App
}

func (c *WhoamiCommand) Help() string {
	helpText := `
Usage: owh whoami

	Shows info about the user currently logged in
`
	return strings.TrimSpace(helpText)
}

func (c *WhoamiCommand) Synopsis() string {
	return "Show info about the user currently logged in"
}

func (c *WhoamiCommand) Run(args []string) int {
	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	me, err := client.GetMe()
	if err != nil {
		return c.View.PrintErr(err)
	}

	rows := []view.LabelValue{
		{Label: "Name", Value: fmt.Sprintf("%s %s", me.FirstName, me.Name)},
		{Label: "ID", Value: me.NicHandle},
	}

	c.View.VerticalTable("", rows)
	return 0
}
