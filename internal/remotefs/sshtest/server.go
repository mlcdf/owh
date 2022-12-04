package sshtest

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	image = "mdns/sshd:latest"

	port     = 2222
	user     = "user"
	password = "password"
	folder   = "www"
)

type containerID struct {
	id string
}

func MustStart() *containerID {
	container, err := Start()
	if err != nil {
		log.Fatalln("fuuuuuuuu", err)
	}
	return container
}

func Start() (*containerID, error) {
	cmd := exec.Command(
		"docker", "run", "-d", "-p", fmt.Sprintf("%d:22", port), "--name", "sftp-root", fmt.Sprintf("-e USER=%s", user), fmt.Sprintf("-e PASSWORD=%s", password), image,
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

func (c containerID) Nuke() {
	if err := c.Kill(); err != nil {
		log.Fatalln("f", err)
	}

	if err := c.Remove(); err != nil {
		log.Fatalln("u", err)
	}
}

func (c containerID) Kill() error {
	return exec.Command("docker", "kill", c.id).Run()
}

func (c containerID) Remove() error {
	return exec.Command("docker", "rm", c.id).Run()
}
