package commands

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
	"golang.org/x/exp/slices"
)

func Users(client *api.Client, hosting string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			return err
		}

		hosting = link.Hosting
	}

	users, err := client.ListUsers(hosting)
	if err != nil {
		return err
	}

	hostingInfo, err := client.GetHosting(hosting)
	if err != nil {
		return err
	}

	tables := make([][]string, 0)

	for _, user := range users {
		var primaryLogin bool

		if hostingInfo.PrimaryLogin == user {
			primaryLogin = true
		}

		if credentials, ok := config.GlobalOpts.SFTPCredentials[hosting]; ok && credentials.User == user {
			user = cmdutil.Special(user)
		}

		row := []string{
			user,
			strconv.FormatBool(primaryLogin),
		}
		tables = append(tables, row)
	}

	return cmdutil.PrintTable("", tables, "Login", "Primary login")
}

func DeleteUser(client *api.Client, hosting string, user string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			return err
		}

		hosting = link.Hosting
	}

	if user == "" {
		if !cmdutil.IsInteractive() {
			fmt.Printf("missing argument --user\n")
			return cmdutil.ErrFlag
		}

		hostingInfo, err := client.GetHosting(hosting)
		if err != nil {
			return err
		}

		users, err := client.ListUsers(hosting)
		if err != nil {
			return err
		}

		if i := slices.Index(users, hostingInfo.PrimaryLogin); i >= 0 {
			users = slices.Delete(users, i, i+1)
		}

		if len(users) == 0 {
			fmt.Println("")
		}

		prompt := &survey.Select{
			Message: "Select user to delete", Options: users}

		err = survey.AskOne(prompt, &user)
		if err != nil {
			return err
		}
	}

	if credentials, ok := config.GlobalOpts.SFTPCredentials[hosting]; ok && credentials.User == user {
		var shouldContinue bool
		prompt := &survey.Confirm{Message: fmt.Sprintf("Warning: the user %s is used by this very projet. Are you sure you want to delete it?", user), Default: false}

		err := survey.AskOne(prompt, &shouldContinue)
		if err != nil {
			return err
		}

		if !shouldContinue {
			return cmdutil.ErrCancel
		}
	}

	err := client.DeleteUser(hosting, user)
	if err != nil {
		return err
	}

	return nil
}

func ChangePassword(client *api.Client, hosting string, user string, password string) error {
	if hosting == "" {
		link, err := config.EnsureLink()
		if err != nil {
			return err
		}
		hosting = link.Hosting
	}

	if password == "" && !cmdutil.IsInteractive() {
		fmt.Println("missing flag --password")
		return cmdutil.ErrFlag
	}

	if user == "" {
		if credential, ok := config.GlobalOpts.SFTPCredentials[hosting]; ok {
			user = credential.User
			fmt.Printf("Changing password for user %s\n", cmdutil.Color(cmdutil.StyleHighlight).Render(user))
		} else {
			fmt.Println("Missing positional argument user")
			return cmdutil.ErrSilent
		}
	}

	return flow.ChangePassword(client, hosting, user, password)
}
