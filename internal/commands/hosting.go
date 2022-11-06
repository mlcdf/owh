package commands

import (
	"fmt"
	"strings"

	"go.mlcdf.fr/owh/internal/api"
	"golang.org/x/xerrors"
)

func Hosting(client *api.Client, hosting string) error {
	var webs []string
	err := client.Get("/hosting/web", &webs)
	if err != nil {
		return xerrors.Errorf("failed to get /hosting/web: %w", err)
	}

	fmt.Println(strings.Join(webs, "\n"))
	return nil
}
