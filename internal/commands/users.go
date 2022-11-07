package commands

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/lipgloss"
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

	hostingInfo, err := client.HostingInfo(hosting)
	if err != nil {
		return err
	}

	style := lipgloss.NewStyle().Width(25)
	for _, user := range users {

		if credentials, ok := config.GlobalOpts.SFTPCredentials[hosting]; ok && credentials.User == user {
			fmt.Printf("%s (configured on this machine)\n", style.Render(user))
		} else if hostingInfo.PrimaryLogin == user {
			fmt.Printf("%s (primary login)\n", style.Render(user))
		} else {
			fmt.Printf("%s\n", user)
		}

	}

	return nil
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
		hostingInfo, err := client.HostingInfo(hosting)
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
