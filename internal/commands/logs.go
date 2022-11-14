package commands

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/sally/cache"
)

func Logs(client *api.Client, c *cache.Cache, homepage bool, owstats bool) error {
	link, err := config.EnsureLink()
	if err != nil {
		return err
	}
	hosting := link.Hosting

	token := c.Get(client.ConsumerKey + "USER_LOGS")
	if token == "" {

		validity := 1 * time.Hour

		token, err := client.GetUserLogsToken(hosting, validity)
		if err != nil {
			return err
		}

		err = c.Set(client.ConsumerKey+"USER_LOGS", token, validity)
		if err != nil {
			return err
		}
	}

	hostingInfo, err := client.GetHosting(hosting)
	if err != nil {
		return err
	}

	if owstats {
		url := owstatsURL(hostingInfo, token)
		err = browser.OpenURL(url)
		if err != nil {
			return err
		}
	} else if homepage {
		url := homeURL(hostingInfo, token)
		err = browser.OpenURL(url)
		if err != nil {
			return err
		}
	} else {
		return lastLogs(hostingInfo, token)
	}

	return nil
}

func lastLogs(hosting *api.HostingInfo, token string) error {
	today := time.Now()

	url := fmt.Sprintf(
		"https://%s/%s/osl/%s-%d-%d-%d.log?token=%s",
		strings.Replace(hosting.ServiceName, hosting.PrimaryLogin, "logs", 1),
		hosting.ServiceName,
		hosting.ServiceName,
		int(today.Day()),
		int(today.Month()),
		int(today.Year()),
		token,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

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
