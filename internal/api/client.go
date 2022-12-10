package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"golang.org/x/xerrors"
)

const applicationKey = "8f13da4094013fac"
const applicationSecret = "13eb34fe75e863d12f97ec62124db9c7"

const defaultTimeout = 30 * time.Second

type ClientFactory func(httpClient *http.Client, region string, consumerKey string) (*Client, error)

type Client struct {
	*ovh.Client
}

var _ ClientFactory = NewClient

func NewClient(httpClient *http.Client, region string, consumerKey string) (*Client, error) {
	client, err := ovh.NewClient(
		region,
		applicationKey,
		applicationSecret,
		consumerKey,
	)

	if err != nil {
		return nil, xerrors.Errorf("failed to instantiate a new ovh Client: %w", err)
	}

	client.Client = httpClient
	client.UserAgent = "owh"
	client.Timeout = defaultTimeout

	if consumerKey == "" {
		return &Client{client}, nil
	}

	apiCredentials := &credentials{}
	err = client.Get("/auth/currentCredential", &apiCredentials)

	if err == nil && apiCredentials.Status == "validated" {
		return &Client{client}, nil
	}

	var e *ovh.APIError
	if errors.As(err, &e) {
		if e.Code == 403 && e.Message == "This credential is not valid" {
			return nil, fmt.Errorf("Current OVHcloud API token is invalid or expired. Please run: owh login")
		}
	}

	return nil, err
}
