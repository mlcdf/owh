package commands

import (
	"fmt"

	"go.mlcdf.fr/owh/internal/api"
)

func Whoami(client *api.Client) error {
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	fmt.Printf("%s %s (%s)\n", me.FirstName, me.Name, me.NicHandle)
	return nil
}
