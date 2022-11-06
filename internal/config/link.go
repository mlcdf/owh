package config

import (
	"fmt"
	"os"
	"path"

	"go.mlcdf.fr/owh/internal/cmdutil"
	"golang.org/x/xerrors"
)

var ErrFolderNotLinked = xerrors.Errorf("folder not linked")

type Link struct {
	// config file location on disk
	location string `json:"-"`

	Hosting         string `json:"hosting,omitempty"`
	CanonicalDomain string `json:"canonical_domain,omitempty"`
}

// EnsureLink retrieve a Link from environment variables and config file
func EnsureLink() (*Link, error) {
	hosting := os.Getenv("OWH_HOSTING")
	canonicalDomain := os.Getenv("OWH_CANONICAL_DOMAIN")

	if hosting == "" && canonicalDomain == "" && !cmdutil.IsInteractive() {
		fmt.Println("Please set the OWH_HOSTING and OWH_CANONICAL_DOMAIN environment variables")
		return nil, cmdutil.ErrSilent
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	location := path.Join(wd, ".owh.json")
	l := &Link{location: location}

	err = fromFile(l, location)
	if err != nil {
		return l, cmdutil.ErrSilent
	}

	if l.Hosting == "" || l.CanonicalDomain == "" {
		return l, ErrFolderNotLinked
	}

	return l, nil
}
