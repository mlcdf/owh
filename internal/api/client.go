package api

import (
	"time"

	"github.com/ovh/go-ovh/ovh"
	"go.mlcdf.fr/owh/internal/config"
	"golang.org/x/xerrors"
)

const applicationKey = "8f13da4094013fac"
const applicationSecret = "13eb34fe75e863d12f97ec62124db9c7"

const defaultTimeout = 2 * time.Minute

type Client struct {
	*ovh.Client
}

// UnloggedClient is a client without the consumer key.
// It should only be used for the login flow.
func NewUnloggedClient(region string) (*Client, error) {
	client, err := ovh.NewClient(
		region,
		applicationKey,
		applicationSecret,
		"",
	)

	client.UserAgent = "owh"
	client.Timeout = defaultTimeout

	if err != nil {
		return nil, xerrors.Errorf("failed to instantiate a new ovh Client: %w", err)
	}

	return &Client{client}, nil
}

func NewClient(region string) (*Client, error) {
	client, err := ovh.NewClient(
		region,
		applicationKey,
		applicationSecret,
		config.GlobalOpts.ConsumerKey,
	)

	client.UserAgent = "owh"
	client.Timeout = defaultTimeout

	if err != nil {
		return nil, xerrors.Errorf("failed to instantiate a new ovh Client: %w", err)
	}

	return &Client{client}, nil
}
