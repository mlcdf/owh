package command

import (
	"flag"
	"strconv"
	"strings"

	"go.mlcdf.fr/owh/internal/cmdutil"
)

type UsersCommand struct {
	App
}

func (c *UsersCommand) Help() string {
	helpText := `
Usage: owh users [--help] <command> [<args>]
	
  Manages users.
`
	return strings.TrimSpace(helpText)
}

func (c *UsersCommand) Synopsis() string {
	return "Manage users"
}

func (c *UsersCommand) Run(args []string) int {
	var hosting string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")

	if err := flags.Parse(args); err != nil {
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

	users, err := client.ListUsers(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	hostingInfo, err := client.GetHosting(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	tables := make([][]string, 0)

	for _, user := range users {
		var primaryLogin bool

		if hostingInfo.PrimaryLogin == user {
			primaryLogin = true
		}

		if credentials, ok := c.Config.SFTPCredentials[hosting]; ok && credentials.User == user {
			user = cmdutil.Special(user)
		}

		row := []string{
			user,
			strconv.FormatBool(primaryLogin),
		}
		tables = append(tables, row)
	}

	err = c.View.Table("", tables, "Login", "Primary login")
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
