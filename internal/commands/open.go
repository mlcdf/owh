package commands

import (
	"fmt"

	"github.com/pkg/browser"
	"go.mlcdf.fr/owh/internal/config"
)

func Open() error {
	link, err := config.EnsureLink()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s", link.CanonicalDomain)

	fmt.Printf("Opening %s ...\n", url)
	return browser.OpenURL(url)
}
