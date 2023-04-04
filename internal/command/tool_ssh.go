package command

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type SSHCommand struct {
	App
}

func (c *SSHCommand) Help() string {
	helpText := `
Usage: owh tool ssh

  Provides a quick way to connect to the linked hosting.

  Since it relies on openssh-client and sshpass, you'll need to have both
  installed.
`
	return strings.TrimSpace(helpText)
}

func (c *SSHCommand) Synopsis() string {
	return "Connect to the linked hosting"
}

func (c *SSHCommand) Run(_ []string) int {
	link, err := c.EnsureLink()
	if err != nil {
		return c.View.PrintErr(err)
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	credentials, ok := c.Config.SFTPCredentials[link.Hosting]
	if !ok {
		c.View.Println("Couldn't find credentials for", link.Hosting)
		return 1
	}

	hosting, err := client.GetHosting(link.Hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	// We relies on ssh & sshpass because I wan't able to get a nice UX while using /x/crypto/ssh.
	// I was able to connect, but had issue with escape sequences (for example arrow up/down to navigate history, etc...)

	dependencies := []string{"ssh", "sshpass"}
	for _, dep := range dependencies {
		_, err := exec.LookPath(dep)
		if err != nil {
			fmt.Printf("You need to have %s installed\n", dep)
			return 0
		}
	}

	args := []string{
		"-p", credentials.Password,
		"ssh", hosting.ServiceManagementAccess.SSH.URL, "-t",
		"-l", credentials.User,
		"-p", fmt.Sprint(hosting.ServiceManagementAccess.SSH.Port),
	}

	cmd := exec.Command("sshpass", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
