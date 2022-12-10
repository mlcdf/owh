package command

import (
	"flag"
	"fmt"
	"strings"

	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/flow"
)

type UsersChangePassCommand struct {
	App
}

func (c *UsersChangePassCommand) Help() string {
	helpText := `
Usage: owh users changepass
	
  Change the password of a ssh/ftp user.
`
	return strings.TrimSpace(helpText)
}

func (c *UsersChangePassCommand) Synopsis() string {
	return "Change ssh/ftp users password"
}

func (c *UsersChangePassCommand) Run(args []string) int {
	var hosting string
	var user string
	var password string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")
	flags.StringVar(&user, "user", "", "")
	flags.StringVar(&password, "password", "", "")

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

	if hosting == "" {
		link, err := c.EnsureLink()
		if err != nil {
			return c.View.PrintErr(err)
		}
		hosting = link.Hosting
	}

	if password == "" && !c.IsInteractive {
		fmt.Println("missing flag --password")
		return 1
	}

	if user == "" {
		if credential, ok := c.Config.SFTPCredentials[hosting]; ok {
			user = credential.User
			fmt.Printf("Changing password for user %s\n", cmdutil.Color(cmdutil.StyleHighlight).Render(user))
		} else {
			fmt.Println("Missing positional argument user")
			return 1
		}
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	err = flow.ChangePassword(client, c.Config, hosting, user, password)
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
