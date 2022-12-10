package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ovh/go-ovh/ovh"
	"github.com/pkg/browser"
	"go.mlcdf.fr/owh/internal/api"
)

type LoginCommand struct {
	App
}

func (c *LoginCommand) Help() string {
	helpText := `
Usage: owh login

  Login to your OVHcloud account.
`
	return strings.TrimSpace(helpText)
}

func (c *LoginCommand) Synopsis() string {
	return "Login to your OVHcloud account"
}

func (c *LoginCommand) Run(args []string) int {
	if c.Config.ConsumerKey != "" && c.IsInteractive {
		var shouldContinue bool

		err := survey.AskOne(
			&survey.Confirm{Message: "You're already logged in. Do you want to re-authenticate?"},
			&shouldContinue,
		)
		if err != nil {
			fmt.Printf("failed to display prompt: %s\n", err)
			return 1
		}

		if !shouldContinue {
			return 2
		}
	}

	var selectedRegion string
	prompt := &survey.Select{Message: "Select a region", Options: []string{"ovh-eu", "ovh-ca"}, Default: "ovh-eu"}

	err := survey.AskOne(prompt, &selectedRegion)
	if err != nil {
		return c.View.PrintErr(err)
	}

	c.Config.Region = selectedRegion

	client, err := api.NewClient(c.HTTPClient, selectedRegion, "")
	if err != nil {
		return c.View.PrintErr(err)
	}

	ckReq := client.NewCkRequest()
	ckReq.AddRules(ovh.ReadOnly, "/me")
	ckReq.AddRecursiveRules(ovh.ReadWrite, "/hosting/web")

	response, err := ckReq.Do()
	if err != nil {
		if err != nil {
			fmt.Printf("error creating consumer key: %v\n", err)
			return 1
		}
	}

	var browserErr error
	if c.IsInteractive {
		browser.Stderr = io.Discard // hide gtk logs on Linux
		browserErr = browser.OpenURL(response.ValidationURL)
	}

	if browserErr != nil {
		fmt.Printf("Please visit %s and validate the form\n", response.ValidationURL)
	}

	err = client.WaitForValidation()
	if err != nil {
		return c.View.PrintErr(err)
	}

	c.Config.ConsumerKey = response.ConsumerKey
	if err = c.Config.Save(); err != nil {
		return c.View.PrintErr(err)
	}

	me, err := client.GetMe()
	if err != nil {
		return c.View.PrintErr(err)
	}

	fmt.Printf("Logged in as %s %s (%s)\n", me.FirstName, me.Name, me.NicHandle)
	return 0
}
