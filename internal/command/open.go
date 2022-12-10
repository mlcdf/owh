package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/browser"
)

type OpenCommand struct {
	App
}

func (c *OpenCommand) Help() string {
	helpText := `
Usage: owh open

  Open browser to current deployed website.
`
	return strings.TrimSpace(helpText)
}

func (c *OpenCommand) Synopsis() string {
	return "Open browser to current deployed website"
}

func (c *OpenCommand) Run(args []string) int {
	link, err := c.EnsureLink()
	if err != nil {
		return c.View.PrintErr(err)
	}

	url := fmt.Sprintf("https://%s", link.CanonicalDomain)

	fmt.Printf("Opening %s ...\n", url)

	browser.Stderr = io.Discard // hide gtk logs on Linux
	err = browser.OpenURL(url)
	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}
