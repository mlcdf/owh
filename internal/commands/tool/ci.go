package tool

import (
	"fmt"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
)

func CI(client *api.Client) error {
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

	return nil
}
