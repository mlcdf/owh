package commands

import (
	"fmt"
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ovh/go-ovh/ovh"
	"github.com/pkg/browser"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"golang.org/x/xerrors"
)

func Login() error {
	if config.GlobalOpts.ConsumerKey != "" && cmdutil.IsInteractive() {
		var shouldContinue bool

		err := survey.AskOne(&survey.Confirm{Message: "You're already logged in. Do you want to re-authenticate?"}, &shouldContinue)
		if err != nil {
			return xerrors.Errorf("failed to display prompt %w", err)
		}

		if !shouldContinue {
			return nil
		}
	}

	var selectedRegion string
	prompt := &survey.Select{Message: "Select a region", Options: []string{"ovh-eu", "ovh-ca"}, Default: "ovh-eu"}

	err := survey.AskOne(prompt, &selectedRegion)
	if err != nil {
		return err
	}

	config.GlobalOpts.Region = selectedRegion

	client, err := api.NewClient(selectedRegion)
	if err != nil {
		return err
	}

	ckReq := client.NewCkRequest()
	ckReq.AddRules(ovh.ReadOnly, "/me")
	ckReq.AddRecursiveRules(ovh.ReadWrite, "/hosting/web")

	response, err := ckReq.Do()
	if err != nil {
		if err != nil {
			return xerrors.Errorf("error creating consumer key: %w", err)
		}
	}

	var browserErr error
	if cmdutil.IsInteractive() {
		browser.Stderr = io.Discard // hide gtk logs on Linux
		browserErr = browser.OpenURL(response.ValidationURL)
	}

	if !cmdutil.IsInteractive() || browserErr != nil {
		fmt.Printf("Please visit %s and validate the form\n", response.ValidationURL)
	}

	err = client.WaitForValidation()
	if err != nil {
		return err
	}

	config.GlobalOpts.ConsumerKey = response.ConsumerKey
	err = config.GlobalOpts.Save()
	if err != nil {
		return err
	}

	me, err := client.GetMe()
	if err != nil {
		return err
	}

	fmt.Printf("Logged in as %s %s (%s)\n", me.FirstName, me.Name, me.NicHandle)
	return nil
}
