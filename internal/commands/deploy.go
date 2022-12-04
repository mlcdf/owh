package commands

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/flow"
	"golang.org/x/xerrors"
)

const tarFilename = "deploy.tar.gz"

type DeployOptions struct {
	// Directory to upload
	Directory string

	// Whether we should attached also attached the www subdomain
	WWW bool
}

func Deploy(client *api.Client, options *DeployOptions) error {
	l, err := config.NewLink()

	if err != nil {
		if err != config.ErrFolderNotLinked {
			return err
		}

		if !cmdutil.IsInteractive() {
			fmt.Printf(
				"Please set the %s and %s environment variables\n",
				config.ENV_OWH_HOSTING,
				config.ENV_OWH_CANONICAL_DOMAIN,
			)
			return cmdutil.ErrSilent
		}

		path, err := os.Getwd()
		if err != nil {
			return err
		}

		prompt := &survey.Confirm{
			Message: fmt.Sprintf("No link found. Set up and deploy %s", path),
			Default: true,
		}
		var shouldContinue bool
		err = survey.AskOne(prompt, &shouldContinue)
		if err != nil {
			return err
		}

		if !shouldContinue {
			return cmdutil.ErrCancel
		}

		err = flow.LinkDirectory(client, l)
		if err != nil {
			return err
		}
	}

	fileToWrite, err := os.OpenFile(tarFilename, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
	if err != nil {
		return xerrors.Errorf("failed to open file %s: %w", tarFilename, err)
	}
	defer fileToWrite.Close()

	conn, err := flow.NewSSHClient(client, l.Hosting)
	if err != nil {
		return xerrors.Errorf("failed to connect ssh: %w", err)
	}

	err = conn.Sync(options.Directory, l.CanonicalDomain)
	if err != nil {
		return err
	}

	fmt.Printf("Files uploaded to ./%s\n", cmdutil.Highlight(l.CanonicalDomain))

	_, err = flow.AttachDomain(client, l.Hosting, l.CanonicalDomain, options.WWW)
	if err != nil {
		return err
	}

	return nil
}
