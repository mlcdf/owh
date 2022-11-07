package commands

import (
	"fmt"

	"github.com/pkg/browser"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
)

func Open() error {
	link, err := config.EnsureLink()
	if err != nil {
		if err != config.ErrFolderNotLinked {
			return err
		}

		fmt.Println("Folder not link. Please run: owh link first")
		return cmdutil.ErrSilent
	}

	url := fmt.Sprintf("https://%s", link.CanonicalDomain)

	fmt.Printf("Opening %s ...\n", url)
	return browser.OpenURL(url)
}
