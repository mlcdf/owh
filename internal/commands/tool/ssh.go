package tool

import (
	"fmt"
	"os"
	"os/exec"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
)

func SSH(client *api.Client) error {
	link, err := config.EnsureLink()
	if err != nil {
		return err
	}

	credentials, ok := config.GlobalOpts.SFTPCredentials[link.Hosting]
	if !ok {
		return nil
	}

	hosting, err := client.GetHosting(link.Hosting)
	if err != nil {
		return err
	}

	// We relies on ssh & sshpass because I wan't able to get a nice UX while using /x/crypto/ssh.
	// I was able to connect, but had issue with escape sequences (for example arrow up/down to navigate history, etc...)

	dependencies := []string{"ssh", "sshpass"}
	for _, dep := range dependencies {
		_, err := exec.LookPath(dep)
		if err != nil {
			fmt.Printf("You need to have %s installed", dep)
			return cmdutil.ErrSilent
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

	return cmd.Run()
}
