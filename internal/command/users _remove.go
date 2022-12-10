package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/exp/slices"
)

type UsersRemoveCommand struct {
	App
}

func (c *UsersRemoveCommand) Help() string {
	helpText := `
Usage: owh users remove
	
  Delete a ssh/ftp user
`
	return strings.TrimSpace(helpText)
}

func (c *UsersRemoveCommand) Synopsis() string {
	return "Remove ssh/ftp users"
}

func (c *UsersRemoveCommand) Run(args []string) int {
	var hosting string
	var user string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&hosting, "hosting", "", "")
	flags.StringVar(&user, "user", "", "")

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

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	if user == "" {
		if !c.IsInteractive {
			fmt.Printf("missing argument --user\n")
			return 1
		}

		hostingInfo, err := client.GetHosting(hosting)
		if err != nil {
			return c.View.PrintErr(err)
		}

		users, err := client.ListUsers(hosting)
		if err != nil {
			return c.View.PrintErr(err)
		}

		if i := slices.Index(users, hostingInfo.PrimaryLogin); i >= 0 {
			users = slices.Delete(users, i, i+1)
		}

		if len(users) == 0 {
			fmt.Println("No user to delete")
			return 1
		}

		prompt := &survey.Select{
			Message: "Select user to delete", Options: users}

		err = survey.AskOne(prompt, &user)
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	if credentials, ok := c.Config.SFTPCredentials[hosting]; ok && credentials.User == user {
		var shouldContinue bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Warning: the user %s is used by this very projet. Are you sure you want to delete it?", user),
			Default: false,
		}

		err := survey.AskOne(prompt, &shouldContinue)
		if err != nil {
			return c.View.PrintErr(err)
		}

		if !shouldContinue {
			return 2
		}
	}

	err = client.DeleteUser(hosting, user)
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
