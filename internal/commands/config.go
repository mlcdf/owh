package commands

import (
	"errors"
	"fmt"

	"go.mlcdf.fr/owh/internal/config"
)

func Config(asEnv bool) error {
	if !asEnv {
		fmt.Println(config.GlobalOpts.Location())
		return nil
	}

	printenv(config.ENV_REGION, config.GlobalOpts.Region)
	printenv(config.ENV_CONSUMER_KEY, config.GlobalOpts.ConsumerKey)

	link, err := config.NewLink()
	if err != nil {
		if !errors.Is(err, config.ErrFolderNotLinked) {
			return err
		}
		return nil
	}

	printenv(config.ENV_OWH_HOSTING, link.Hosting)
	printenv(config.ENV_OWH_CANONICAL_DOMAIN, link.CanonicalDomain)

	if credentials, ok := config.GlobalOpts.SFTPCredentials[link.Hosting]; ok {
		printenv(config.ENV_SSH_USER, credentials.User)
		printenv(config.ENV_SSH_PASSWORD, credentials.Password)
	}

	return nil
}

func printenv(label, value string) {
	fmt.Printf("%s=\"%s\"\n", label, value)
}
