package command

import (
	"fmt"
	"strings"
)

type CICommand struct {
	App
}

func (c *CICommand) Help() string {
	helpText := `
Usage: owh tool ci

  Help you setup a deployment in CI.

  /!\ This command will display secrets in the terminal.
`
	return strings.TrimSpace(helpText)
}

func (c *CICommand) Synopsis() string {
	return "Shows useful info to setup a CI"
}

func (c *CICommand) Run(args []string) int {
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
		return 1
	}

	hosting, err := client.GetHosting(link.Hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	fmt.Printf(`# Define these 2 secrets in your CI
# OVH_HOSTING_PASSWORD = %s
# OVH_HOSTING_USER = %s

sshpass -p "$OVH_HOSTING_PASSWORD" rsync -av --exclude '.*' --exclude 'LICENSE' -e "ssh -o StrictHostKeyChecking=no" . $OVH_HOSTING_USER@$%s:~/%s
`,
		credentials.User,
		credentials.Password,
		hosting.ServiceManagementAccess.SSH.URL,
		link.CanonicalDomain,
	)

	return 0
}
