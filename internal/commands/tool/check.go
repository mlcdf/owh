package tool

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"text/tabwriter"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/config"
)

func Check(client *api.Client) error {
	link, err := config.EnsureLink()
	if err != nil {
		return err
	}

	ips, err := net.LookupIP(link.CanonicalDomain)
	if err != nil {
		return err
	}

	// cname, err := net.LookupCNAME(link.CanonicalDomain)
	// if err != nil {
	// 	return err
	// }

	hosting, err := client.GetHosting(link.Hosting)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	isV4OK := "no"
	isV6OK := "no"

	for _, ip := range ips {
		_ip := ip.To4()

		if _ip != nil {
			if _ip.String() == hosting.HostingIP {
				isV4OK = "yes"
			}
		} else {
			if _ip.String() == hosting.HostingIPv6 {
				isV6OK = "yes"
			}
		}
	}

	fmt.Println(cmdutil.Bold("DNS"))

	fmt.Fprintf(w, "  Record A    \t=\t%s\t\n", isV4OK)
	fmt.Fprintf(w, "  Record AAAA \t=\t%s\n", isV6OK)

	w.Flush()

	url := fmt.Sprintf("http://%s", link.CanonicalDomain)

	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	res, err := noRedirectClient.Get(url)
	if err != nil {
		return err
	}

	res.Body.Close()

	fmt.Println()
	fmt.Println(cmdutil.Bold("HTTP"))

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", link.CanonicalDomain), nil)
	if err == nil {
		err = conn.VerifyHostname(link.CanonicalDomain)
		if err != nil {
			fmt.Fprintln(w, "  Valid certificate\t=\tno, hostname doesn't match: %s", err)
		}
		fmt.Fprintln(w, "  Valid certificate\t=\tyes")
	}

	if err != nil {
		fmt.Fprintln(w, "  Valid certificate\t=\tno, %s", err)
	}

	if (res.StatusCode == 301 || res.StatusCode == 302) &&
		res.Header.Get("location") == fmt.Sprintf("https://%s", link.CanonicalDomain) {
		fmt.Fprintln(w, "  Enforce HTTPS\t=\tyes")
	} else {
		fmt.Fprintln(w, "  Enforce HTTPS\t=\tno")
	}

	fmt.Fprintln(w, "  Custom 404\t=\tno")
	fmt.Fprintln(w, "  Custom 505\t=\tno")

	res, err = http.Get("https://" + link.CanonicalDomain)
	if err != nil || res.Request.ProtoMajor != 200 {
		fmt.Fprintln(w, "  HTTP/2\t=\tno")
	} else {
		fmt.Fprintln(w, "  HTTP/2\t=\tyes")
	}

	res.Body.Close()

	w.Flush()

	return nil
}

// func checkCommonPath(url string) error {
// 	paths := []string{
// 		".env",
// 		".owh.json",
// 		".git",
// 	}

// 	for

// 	return nil
// }
