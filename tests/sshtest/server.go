package sshtest

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

const (
	image    = "ghcr.io/mlcdf/sftptest:latest"
	user     = "owh"
	password = "test"
	Port     = 2222
)

var pk ssh.Signer

type containerID struct {
	id string
}

func MustStart(t *testing.T) *containerID {
	t.Helper()

	container, err := Start(t)
	if err != nil {
		t.Fatal("fuuuuuuuu", err)
	}
	return container
}

func Start(t *testing.T) (*containerID, error) {
	cmd := exec.Command(
		"docker", "run", "-d",
		"-p", fmt.Sprintf("%d:22", Port),
		"--name", "sftp-root",
		image, fmt.Sprintf("%s:%s:1001", user, password),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %w", stderr.String(), err)
	}

	id := strings.TrimSpace(stdout.String())
	if id == "" {
		return nil, errors.New("unexpected empty output from `docker run`")
	}

	return &containerID{id}, nil
}

func (c containerID) Nuke(t *testing.T) {
	t.Helper()

	output, _ := c.Logs()
	t.Log(string(output))

	if output, err := c.Kill(); err != nil {
		t.Log(err, string(output))
	}

	if output, err := c.Remove(); err != nil {
		t.Log(err, string(output))
	}
}

func (c containerID) Kill() ([]byte, error) {
	return exec.Command("docker", "kill", c.id).CombinedOutput()
}

func (c containerID) Remove() ([]byte, error) {
	return exec.Command("docker", "rm", c.id).CombinedOutput()
}

func (c containerID) Logs() ([]byte, error) {
	return exec.Command("docker", "logs", c.id).CombinedOutput()
}
