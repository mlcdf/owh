package command

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
)

type CheckCommand struct {
	App
}

func (c *CheckCommand) Help() string {
	helpText := `
Usage: owh tool check

  Performs various check on your website such as:
	- check DNS config
	- validate SSL certs
	- test http => https redirection
	- etc
`
	return strings.TrimSpace(helpText)
}

func (c *CheckCommand) Synopsis() string {
	return "Perform various check on your website"
}

func (c *CheckCommand) Run(args []string) int {
	var hosting string
	var domain string

	flags := flag.NewFlagSet("link", flag.ExitOnError)

	flags.StringVar(&domain, "domain", "", "")

	if err := flags.Parse(args); err != nil {
		return c.View.PrintErr(err)
	}

	if domain == "" {
		link, err := c.EnsureLink()
		if err != nil && err != config.ErrFolderNotLinked {
			return c.View.PrintErr(err)
		}

		domain = link.CanonicalDomain
		hosting = link.Hosting
	}

	client, err := c.LoggedClient()
	if err != nil {
		return c.View.PrintErr(err)
	}

	if hosting == "" {
		hosting, err = client.HostingByDomain(domain)
		if err != nil {
			return c.View.PrintErr(err)
		}
	}

	ch := &check{wg: new(sync.WaitGroup), httpClient: c.HTTPClient}
	ch.wg.Add(5)

	go ch.checkEnforceHTTPS(domain)
	go ch.checkValidCert(domain)
	go ch.checkProtocol(domain)
	go ch.checCustom404(domain)

	hostingInfo, err := client.GetHosting(hosting)
	if err != nil {
		return c.View.PrintErr(err)
	}

	go ch.checkIP(hostingInfo, domain)

	ch.wg.Wait()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, cmdutil.Bold("DNS")+"\t\t\n")
	fmt.Fprintf(w, "  Record A    \t=\t%s\n", ch.ipv4)
	fmt.Fprintf(w, "  Record AAAA \t=\t%s\n", ch.ipv6)
	fmt.Fprintf(w, "\t\t\n")
	fmt.Fprintf(w, cmdutil.Bold("HTTP")+"\t\t\n")

	fmt.Fprintf(w, "  Protocol\t=\t%s\n", ch.protocol)
	fmt.Fprintf(w, "  Valid certificate\t=\t%s\n", ch.validCert)
	fmt.Fprintf(w, "  Enforce HTTPS\t=\t%s\n", ch.enforceHTTPS)
	fmt.Fprintf(w, "  Custom 404 page\t=\t%s\n", ch.custom404)

	return 0
}

type check struct {
	wg         *sync.WaitGroup
	httpClient *http.Client

	ipv4 string
	ipv6 string

	protocol     string
	validCert    string
	enforceHTTPS string
	custom404    string
}

func (c *check) checkIP(hosting *api.HostingInfo, domain string) {
	defer c.wg.Done()

	ips, err := net.LookupIP(domain)
	if err != nil {
		c.ipv4 = "failed to lookup IP"
		c.ipv6 = "failed to lookup IP"
		return
	}

	var ipv4 bool
	var ipv6 bool

	for _, ip := range ips {
		_ip := ip.To4()

		if _ip != nil {
			if _ip.String() == hosting.HostingIP {
				ipv4 = true
			}
		} else {
			if ip.To16().String() == hosting.HostingIPv6 {
				ipv6 = true
			}
		}
	}

	c.ipv4 = yesno(ipv4)
	c.ipv6 = yesno(ipv6)
}

func yesno(value bool) string {
	if value {
		return "yes"
	}

	return "no"
}

func (c *check) checkValidCert(domain string) {
	defer c.wg.Done()

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), nil)
	if err == nil {
		err = conn.VerifyHostname(domain)
		if err != nil {
			c.validCert = fmt.Sprintf("no, hostname doesn't match: %s", err)
		}

		c.validCert = "yes"
		return
	}

	c.validCert = fmt.Sprintf("no, %s", err)
}

func (c *check) checkEnforceHTTPS(domain string) {
	defer c.wg.Done()

	url := fmt.Sprintf("http://%s", domain)

	noRedirectClient := c.httpClient
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error { return nil }

	res, err := noRedirectClient.Get(url)
	if err != nil {
		c.enforceHTTPS = err.Error()
		return
	}

	res.Body.Close()

	if (res.StatusCode == 301 || res.StatusCode == 302) &&
		strings.Contains(res.Header.Get("location"), fmt.Sprintf("https://%s", domain)) {
		c.enforceHTTPS = "yes"
		return
	}

	c.enforceHTTPS = "no"
}

func (c *check) checkProtocol(domain string) {
	defer c.wg.Done()

	res, err := c.httpClient.Get("https://" + domain)
	if err != nil {
		c.protocol = err.Error()
		return
	}
	res.Body.Close()

	c.protocol = res.Proto
}

func (c *check) checCustom404(domain string) {
	defer c.wg.Done()

	res, err := c.httpClient.Get(fmt.Sprintf("https://%s/thispagedoesnotexists", domain))
	if err != nil {
		c.custom404 = err.Error()
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.custom404 = err.Error()
		return
	}

	if strings.Contains(string(body), "<title>404 Not Found</title>") &&
		strings.Contains(string(body), "<p>The requested URL was not found on this server.</p>") {
		c.custom404 = "no"
		return
	}

	c.custom404 = "yes"
}
