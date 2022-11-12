package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"golang.org/x/xerrors"
)

type credentials struct {
	Status string `json:"status"`
}

func (client *Client) WaitForValidation() error {
	var retry int

	for retry < 60 {
		apiCredentials := &credentials{}
		err := client.Get("/auth/currentCredential", &apiCredentials)

		if err == nil && apiCredentials.Status == "validated" {
			return nil
		}

		var e *ovh.APIError
		if errors.As(err, &e) {
			if e.Code == http.StatusForbidden && e.Message == "This credential is not valid" {
				time.Sleep(2 * time.Second)
				retry++
			}
		}
	}

	return xerrors.New("timed out waiting for authorization")
}
