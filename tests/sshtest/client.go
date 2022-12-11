package sshtest

import (
	"golang.org/x/crypto/ssh"
)

func SSHKeyConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}
