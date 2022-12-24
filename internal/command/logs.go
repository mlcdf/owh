package command

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/sally/cache"
)

type LogsCommand struct {
	App

	CacheFactory func() (cache.Cache, error)
}

func (c *LogsCommand) Help() string {
	helpText := `
Usage: owh logs

  View access logs.

Options:
  --owstats        open OVHcloud Web Statistics
  --homepage       open the logs homepage
`
	return strings.TrimSpace(helpText)
}

func (c *LogsCommand) Synopsis() string {
	return "View access logs"
}

func (c *LogsCommand) Run(args []string) int {
	var owstats bool
	var homepage bool

	flags := flag.NewFlagSet("logs", flag.ExitOnError)

	flags.BoolVar(&owstats, "owstats", false, "")
	flags.BoolVar(&homepage, "homepage", false, "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	if owstats && homepage {
		fmt.Println("Flags owstats and homepage can't be step at the same time.")
		return 1
	}

	link, err := c.EnsureLink()
	if err != nil {
		return c.View.PrintErr(err)
	}

	hosting := link.Hosting

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	cache, err := c.CacheFactory()
	if err != nil {
		return c.View.PrintErr(err)
	}

	token := cache.Get(c.Config.ConsumerKey + "USER_LOGS")
	if token == "" {
		validity := 1 * time.Hour

		token, err = client.GetUserLogsToken(hosting, validity)
		if err != nil {
			return c.View.PrintErr(err)
		}

		err = cache.Set(client.ConsumerKey+"USER_LOGS", token, validity)
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	hostingInfo, err := client.GetHosting(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	switch {
	case owstats:
		url := owstatsURL(hostingInfo, token)
		browser.Stderr = io.Discard // hide gtk logs on Linux
		err = browser.OpenURL(url)
	case homepage:
		url := homeURL(hostingInfo, token)
		browser.Stderr = io.Discard // hide gtk logs on Linux
		err = browser.OpenURL(url)
	default:
		err = lastLogs(c.HTTPClient, hostingInfo, token)
	}

	if err != nil {
		return c.View.PrintErr(err)
	}

	return 0
}

func lastLogs(httpClient *http.Client, hosting *api.HostingInfo, token string) error {
	today := time.Now()

	url := fmt.Sprintf(
		"https://%s/%s/osl/%s-%d-%d-%d.log?token=%s",
		strings.Replace(hosting.ServiceName, hosting.PrimaryLogin, "logs", 1),
		hosting.ServiceName,
		hosting.ServiceName,
		today.Day(),
		int(today.Month()),
		today.Year(),
		token,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fmt.Println(res.Request.URL)

	logs, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Print(string(logs))
	return nil
}

func homeURL(hosting *api.HostingInfo, token string) string {
	return fmt.Sprintf(
		"https://%s/%s?token=%s",
		strings.Replace(hosting.ServiceName, hosting.PrimaryLogin, "logs", 1),
		hosting.ServiceName,
		token,
	)
}

func owstatsURL(hosting *api.HostingInfo, token string) string {
	return fmt.Sprintf(
		"https://%s/%s/owstats?token=%s",
		strings.Replace(hosting.ServiceName, hosting.PrimaryLogin, "logs", 1),
		hosting.ServiceName,
		token,
	)
}
