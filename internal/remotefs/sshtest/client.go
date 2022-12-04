package sshtest

import (
	"go.mlcdf.fr/owh/internal/remotefs"
)

func Connect() (*remotefs.Client, error) {
	return remotefs.Connect("localhost", port, user, password)
}
