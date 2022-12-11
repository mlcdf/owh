package remote

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"go.mlcdf.fr/sally/logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/xerrors"
)

var ErrEmptyStringSrc = errors.New("src cannot be an empty string")
var ErrEmptyStringDest = errors.New("dest cannot be an empty string")

type Client struct {
	conn *ssh.Client
}

type Config struct {
	Host string
	Port int

	SSHConfig *ssh.ClientConfig
}

type ConfigFactory func(host string, port int, user string, password string) *Config

// PasswordConfig returns a password based ssh.ClientConfig.
func NewPasswordConfig(host string, port int, user string, password string) *Config {
	return &Config{
		Host: host,
		Port: port,
		SSHConfig: &ssh.ClientConfig{
			Config:          ssh.Config{},
			User:            user,
			Auth:            []ssh.AuthMethod{ssh.RetryableAuthMethod(ssh.Password(password), 5)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		},
	}
}

var _ ConfigFactory = NewPasswordConfig

func Connect(config *Config) (*Client, error) {
	client := &Client{}

	var err error
	var retry int

	for retry < 5 {
		client.conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), config.SSHConfig)
		if err == nil {
			return client, nil
		}

		retry++
		time.Sleep(1 * time.Second)
	}

	return nil, err
}

func (c *Client) Sync(src string, dest string) error {
	if src == "" {
		return ErrEmptyStringSrc
	}

	if dest == "" {
		return ErrEmptyStringDest
	}

	client, err := sftp.NewClient(c.conn)
	if err != nil {
		return xerrors.Errorf("error creating %s directory %w", dest, err)
	}

	err = client.MkdirAll(dest)
	if err != nil {
		return xerrors.Errorf("error creating %s directory %w", dest, err)
	}

	// Delete extra files
	walker := client.Walk(dest)
	for walker.Step() {
		relpath, err := filepath.Rel(dest, walker.Path())

		if err != nil {
			return err
		}

		localpath := filepath.Join(src, relpath)

		remotepath := filepath.Join(dest, relpath)

		remotefile, err := client.Stat(remotepath)
		if err != nil {
			if err.Error() == "file does not exist" {
				continue
			}
			return xerrors.Errorf("error while stat remote file %s: %w", remotepath, err)
		}

		fmt.Println(localpath)

		localfile, err := os.Stat(localpath)
		if err != nil {
			if os.IsNotExist(err) {
				// The file is present on remote but not locally
				err := c.ForceRemove(remotepath)
				if err != nil {
					return err
				}
				continue
			}
			return xerrors.Errorf("error while stat %s: %w", localpath, err)
		}

		// Both are directories
		if localfile.IsDir() && remotefile.IsDir() {
			continue
		}

		// Both are files
		if !localfile.IsDir() && !remotefile.IsDir() {
			if localfile.Size() == remotefile.Size() {
				identical, err := isIdentical(client, localpath, remotepath)
				if err != nil {
					return err
				}

				if identical {
					continue
				}
			}
		}

		logging.Debugf("%s", remotefile.IsDir())
		// files not identical
		// or one is a dir and the other is a file
		err = c.ForceRemove(remotepath)
		if err != nil {
			return err
		}
	}

	// Create new files
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if path == src {
			return nil
		}

		if err != nil {
			return xerrors.Errorf("error walking %s path %s: %w", src, path, err)
		}

		logging.Debugf("path: %s", path)

		if skipFile(path) {
			logging.Debugf("Path %s skipped", path)
			return nil
		}

		relpath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		remotepath := filepath.Join(dest, relpath)

		localfile, err := os.Stat(path)
		if err != nil {
			return xerrors.Errorf("error while stat %s: %w", path, err)
		}

		if localfile.IsDir() {
			if err := client.MkdirAll(remotepath); err != nil {
				return xerrors.Errorf("error while mkdir %s on remote: %w", remotepath, err)
			}
			return nil
		}

		logging.Debugf(path)

		return createFile(client, path, remotepath)
	})
}

func (c *Client) Run(cmd string) (string, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Println(string(output))
		return "", xerrors.Errorf("failed to run command %s: %w", cmd, err)
	}

	return string(output), nil
}

// ForceRemove performs a rm -rf of the dest.
func (c *Client) ForceRemove(dest string) error {
	_, err := c.Run(fmt.Sprintf("rm -rf %s", dest))
	if err != nil {
		xerrors.Errorf("failed to force remove: %w", err)
	}
	return nil
}

func isIdentical(client *sftp.Client, path1, path2 string) (bool, error) {
	buffer := make([]byte, 10_000_000)
	h1 := md5.New()
	h2 := md5.New()

	f1, err := os.Open(path1)
	if err != nil {
		return false, xerrors.Errorf("error opening file %s: %w", path1, err)
	}
	defer f1.Close()

	f2, err := client.Open(path2)
	if err != nil {
		return false, xerrors.Errorf("error opening file %s: %w", path2, err)
	}

	defer f2.Close()

	if _, err := io.CopyBuffer(h1, f1, buffer); err != nil {
		return false, err
	}

	if _, err := io.CopyBuffer(h2, f2, buffer); err != nil {
		return false, err
	}

	if hex.EncodeToString(h1.Sum(nil)) == hex.EncodeToString(h2.Sum(nil)) {
		return true, nil
	}

	return false, nil
}

func createFile(client *sftp.Client, localpath, remotepath string) error {
	remotef, err := client.Create(remotepath)
	if err != nil {
		return xerrors.Errorf("error creating %s: %w", remotepath, err)
	}

	localf, err := os.Open(localpath)
	if err != nil {
		return xerrors.Errorf("error opening %s: %w", localpath, err)
	}

	_, err = io.Copy(remotef, localf)
	if err != nil {
		return xerrors.Errorf("error copying %s to %s: %w", localpath, remotepath, err)
	}

	logging.Debugf(remotepath)

	return nil
}

func skipFile(path string) bool {
	basepath := filepath.Base(path)

	if strings.Contains(basepath, ".well-known") ||
		strings.Contains(basepath, ".htaccess") {
		return false
	}

	if strings.HasPrefix(path, ".") {
		return true
	}

	if strings.Contains(path, "node_modules") || strings.Contains(path, "__py_cache__") {
		return true
	}

	return false
}
