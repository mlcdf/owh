package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/alitto/pond"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"
)

var NotHostingFound = errors.New("Not hosting found")
var MoreThanOneHostingFound = errors.New("More than one hosting found")

type UnitValue struct {
	Unit  string  `json:"unit"`
	Value float32 `json:"value"`
}

func (uv *UnitValue) String() string {
	return fmt.Sprintf("%.1f %s", uv.Value, uv.Unit)
}

type HostingInfo struct {
	ServiceName             string `json:"serviceName"`
	DisplayName             string `json:"displayName"`
	HasCDN                  bool   `json:"hasCdn"`
	HostingIp               string `json:"hostingIp"`
	HostingIpv6             string `json:"hostingIpv6"`
	State                   string `json:"state"`
	PrimaryLogin            string `json:"primaryLogin"`
	ServiceManagementAccess struct {
		SSH struct {
			Port int    `json:"port"`
			URL  string `json:"url"`
		} `json:"ssh"`
	} `json:"serviceManagementAccess"`
	Offer     string    `json:"offer"`
	QuotaSize UnitValue `json:"quotaSize"`
	QuotaUsed UnitValue `json:"quotaUsed"`
}

type SSHUser struct {
	Home     string `json:"home,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	SSHState string `json:"sshState,omitempty"`
}

type AttachedDomain struct {
	Domain   string `json:"domain"`
	Firewall string `json:"firewall"`
	Path     string `json:"path"`
	SSL      bool   `json:"ssl"`
}

func (client *Client) GetHosting(hosting string) (*HostingInfo, error) {
	var hostingInfo HostingInfo

	err := client.Get("/hosting/web/"+hosting, &hostingInfo)
	if err != nil {
		return nil, err
	}

	return &hostingInfo, nil
}

func (client *Client) ListHostings() ([]string, error) {
	var webs []string
	err := client.Get("/hosting/web", &webs)
	if err != nil {
		return nil, xerrors.Errorf("failed to get /hosting/web: %w", err)
	}
	return webs, nil
}

func (client *Client) HostingByDomain(domain string) (string, error) {
	var hostings []string
	url := fmt.Sprintf("/hosting/web/attachedDomain?domain=%s", domain)

	err := client.Get(url, &hostings)
	if err != nil {
		return "", xerrors.Errorf("failed to get %s: %w", url, err)
	}

	if len(hostings) == 0 {
		return "", NotHostingFound
	}

	if len(hostings) > 1 {
		return "", MoreThanOneHostingFound
	}

	return hostings[0], nil
}

func (client *Client) Hostings() ([]HostingInfo, error) {
	hostings, err := client.ListHostings()
	if err != nil {
		return nil, err
	}

	var hs []HostingInfo
	pool := pond.New(20, 20)
	defer pool.StopAndWait()

	// Create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	for _, hosting := range hostings {
		url := fmt.Sprintf("/hosting/web/%s", hosting)

		group.Submit(func() error {
			var d HostingInfo
			err := client.Get(url, &d)
			if err != nil {
				return err
			}

			hs = append(hs, d)

			return nil
		})
	}

	// Wait for all HTTP requests to complete.
	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return hs, nil
}

func (client *Client) ListDomains(hosting string) ([]string, error) {
	var response []string
	err := client.Get(fmt.Sprintf("/hosting/web/%s/attachedDomain", hosting), &response)
	if err != nil {
		return nil, xerrors.Errorf("failed to get /hosting/web/%s/attachedDomain: %w", hosting, err)
	}

	// Delete the default domain whoes ame match the hosting
	if i := slices.Index(response, hosting); i >= 0 {
		response = slices.Delete(response, i, i+1)
		slices.Sort(response)
	}

	return response, nil
}

func (client *Client) Domains(hosting string) ([]AttachedDomain, error) {
	domains, err := client.ListDomains(hosting)
	if err != nil {
		return nil, err
	}

	var ds []AttachedDomain
	pool := pond.New(20, 20)
	defer pool.StopAndWait()

	// Create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	for _, domain := range domains {
		url := fmt.Sprintf("/hosting/web/%s/attachedDomain/%s", hosting, domain)

		group.Submit(func() error {
			var d AttachedDomain
			err := client.Get(url, &d)
			if err != nil {
				return err
			}

			ds = append(ds, d)

			return nil
		})
	}

	// Wait for all HTTP requests to complete.
	err = group.Wait()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(ds, func(a AttachedDomain, b AttachedDomain) bool {
		return a.Path < b.Path
	})

	return ds, nil
}

func (client *Client) GetDomain(hosting string, domain string) (*AttachedDomain, error) {
	var attachedDomain *AttachedDomain

	err := client.Get(fmt.Sprintf("/hosting/web/%s/attachedDomain/%s", hosting, domain), &attachedDomain)
	if err != nil {
		return nil, xerrors.Errorf("failed to GET /hosting/web/%s/attachedDomain/%s: %w", hosting, domain, err)
	}

	return attachedDomain, nil
}

func (client *Client) UpdateDomain(hosting string, domain string) error {
	attachedDomain := &AttachedDomain{
		Domain:   domain,
		Firewall: "active",
		Path:     domain,
		SSL:      true,
	}

	err := client.Put(fmt.Sprintf("/hosting/web/%s/attachedDomain/%s", hosting, domain), attachedDomain, nil)
	if err != nil {
		return xerrors.Errorf("failed to PUT /hosting/web/%s/attachedDomain/%s: %w", hosting, domain, err)
	}

	return nil
}

func (client *Client) PostDomain(hosting string, domain string) error {
	var task *Task

	attachedDomain := &AttachedDomain{
		Domain:   domain,
		Firewall: "active",
		Path:     domain,
		SSL:      true,
	}

	err := client.Post(fmt.Sprintf("/hosting/web/%s/attachedDomain", hosting), attachedDomain, &task)
	if err != nil {
		return xerrors.Errorf("failed to POST /hosting/web/%s/attachedDomain %s: %w", hosting, domain, err)
	}

	err = client.WaitTaskDone(hosting, task.ID)
	if err != nil {
		return xerrors.Errorf("failed to attach domain: %w", err)
	}

	return nil
}

func (client *Client) DeleteDomain(hosting string, domain string) error {
	url := fmt.Sprintf("/hosting/web/%s/attachedDomain/%s", hosting, domain)

	var task Task
	err := client.Delete(url, &task)
	if err != nil {
		return xerrors.Errorf("failed to DELETE %s: %w", url, err)
	}

	err = client.WaitTaskDone(hosting, task.ID)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) ListUsers(hosting string) ([]string, error) {
	var users []string
	url := fmt.Sprintf("/hosting/web/%s/user", hosting)

	err := client.Get(url, &users)
	if err != nil {
		return nil, xerrors.Errorf("failed to GET %s: %w", url, err)
	}

	slices.Sort(users)

	return users, nil
}

func (client *Client) DeleteUser(hosting string, user string) error {
	url := fmt.Sprintf("/hosting/web/%s/user/%s", hosting, user)

	err := client.Delete(url, nil)
	if err != nil {
		return xerrors.Errorf("failed to DELETE %s: %w", url, err)
	}

	return nil
}

func (client *Client) ChangePassword(hosting string, user string, password string) error {
	url := fmt.Sprintf("/hosting/web/%s/user/%s/changePassword", hosting, user)
	payload := &SSHUser{
		Password: password,
	}
	task := &Task{}

	err := client.Post(url, payload, task)
	if err != nil {
		return xerrors.Errorf("failed to POST %s: %w", url, err)
	}

	err = client.WaitTaskDone(hosting, task.ID)
	if err != nil {
		return err
	}

	return nil
}
