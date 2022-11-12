package config

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.mlcdf.fr/owh/internal/cmdutil"
	"golang.org/x/xerrors"
)

const ENV_OWH_HOSTING = ENV_PREFIX + "HOSTING"
const ENV_OWH_CANONICAL_DOMAIN = ENV_PREFIX + "CANONICAL_DOMAIN"

var ErrFolderNotLinked = xerrors.Errorf("Directory not linked")

type Link struct {
	// config file location on disk
	location string `json:"-"`

	Hosting         string `json:"hosting,omitempty"`
	CanonicalDomain string `json:"canonical_domain,omitempty"`
}

func NewLink() (*Link, error) {
	link := &Link{location: path.Join(".owh.json")}

	err := fromFile(link, link.location)
	if err != nil {
		return link, cmdutil.ErrSilent
	}

	if hosting := os.Getenv(ENV_OWH_HOSTING); hosting != "" {
		link.Hosting = hosting
	}

	if canonicalDomain := os.Getenv(ENV_OWH_CANONICAL_DOMAIN); canonicalDomain != "" {
		link.CanonicalDomain = canonicalDomain
	}

	if link.Hosting == "" || link.CanonicalDomain == "" {
		return link, ErrFolderNotLinked
	}

	return link, nil
}

// EnsureLink retrieve a Link from environment variables and config file.
func EnsureLink() (*Link, error) {
	link, err := NewLink()
	if err != nil {
		if err == ErrFolderNotLinked {
			if cmdutil.IsInteractive() {
				fmt.Println("Directory not linked. Please run: owh link")
			} else {
				fmt.Printf("Please set the %s and %s environment variables\n", ENV_OWH_HOSTING, ENV_OWH_CANONICAL_DOMAIN)
			}
			return nil, cmdutil.ErrSilent
		}

		return link, err
	}

	return link, nil
}

func (link *Link) Save() error {
	if err := save(link); err != nil {
		return err
	}

	return link.gitignore()
}

func (link *Link) gitignore() error {
	f, err := os.Open(".gitignore")
	if err != nil {
		return err
	}

	filename := filepath.Base(link.location)

	lines := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), filename) {
			return nil
		}

		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	f.Close()

	lines = append(lines, "\n"+filename)

	err = os.WriteFile(".gitignore", []byte(strings.Join(lines, "\n")), 0640)
	if err != nil {
		return err
	}

	return nil
}
