package api

import (
	"golang.org/x/xerrors"
)

type Me struct {
	Name      string `json:"name"`
	FirstName string `json:"firstname"`
	NicHandle string `json:"nichandle"`
}

func (client *Client) GetMe() (*Me, error) {
	var me Me

	err := client.Get("/me", &me)
	if err != nil {
		return nil, xerrors.Errorf("failed to get /me: %w", err)
	}

	return &me, nil
}
