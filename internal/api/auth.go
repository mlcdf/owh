package api

import (
	"errors"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"golang.org/x/xerrors"
)

type APICredentials struct {
	Status string `json:"status"`
}

func (client *Client) WaitForValidation() error {
	var retry int

	for retry < 60 {
		apiCredentials := &APICredentials{}
		err := client.Get("/auth/currentCredential", &apiCredentials)

		if err == nil && apiCredentials.Status == "validated" {
			return nil
		}

		var e *ovh.APIError
		if errors.As(err, &e) {
			if e.Code == 403 && e.Message == "This credential is not valid" {
				time.Sleep(2 * time.Second)
				retry++
			}
		}
	}

	return xerrors.New("timed out waiting for authorization")
}
