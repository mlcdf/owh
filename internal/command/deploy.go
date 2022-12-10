package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
)

type DeployCommand struct {
	App
}

func (c *DeployCommand) Help() string {
	helpText := `
Usage: owh deploy [options] DIR

  Deploys the linked website to OVHcloud Web Hosting.
  If the directory is not linked, it'll ask to linked it to a hosting first.

Options:
  --www       If present, also attach www/non-www domain
`
	return strings.TrimSpace(helpText)
}

func (c *DeployCommand) Synopsis() string {
	return "Deploy websites from a directory"
}

func (c *DeployCommand) Run(args []string) int {
	var www bool
	var directory string

	flags := flag.NewFlagSet("deploy", flag.ExitOnError)

	flags.BoolVar(&www, "www", false, "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	if flags.Arg(0) != "" {
		directory = flags.Arg(0)
	} else {
		var err error

		directory, err = os.Getwd()
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	ovhapi, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	fmt.Printf("Deploying %s\n", directory)

	l, err := c.App.EnsureLink()

	if err != nil {
		if !errors.Is(err, config.ErrFolderNotLinked) {
			return c.View.PrintErr(err)
		}

		if !c.IsInteractive {
			fmt.Printf(
				"Please set the %s and %s environment variables\n",
				config.ENV_OWH_HOSTING,
				config.ENV_OWH_CANONICAL_DOMAIN,
			)
			return 1
		}

		path, err := os.Getwd()
		if err != nil {
			return c.View.PrintErr(err)
		}

		prompt := &survey.Confirm{
			Message: fmt.Sprintf("No link found. Set up and deploy %s", path),
			Default: true,
		}
		var shouldContinue bool
		err = survey.AskOne(prompt, &shouldContinue)
		if err != nil {
			return c.View.PrintErr(err)
		}

		if !shouldContinue {
			return 2
		}

		err = flow.LinkDirectory(ovhapi, l)
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	conn, err := flow.NewSSHClient(ovhapi, c.Config, c.View, c.IsInteractive, l.Hosting)
	if err != nil {
		fmt.Printf("failed to connect ssh: %v\n", err)
		return 1
	}

	err = conn.Sync(directory, l.CanonicalDomain)
	if err != nil {
		fmt.Printf("failed to upload files: %v\n", err)
		return 1
	}

	fmt.Printf("Files uploaded to ./%s\n", cmdutil.Highlight(l.CanonicalDomain))

	_, err = flow.AttachDomain(ovhapi, l.Hosting, l.CanonicalDomain, www)
	if err != nil {
		fmt.Printf("failed to attach domain: %v\n", err)
		return 1
	}

	return 0
}
