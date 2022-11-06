package flow

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"go.mlcdf.fr/sally/logging"
	"golang.org/x/xerrors"
)

const mega = 1_000_000

func skipFile(path string) bool {
	basepath := filepath.Base(path)

	if strings.HasPrefix(basepath, ".") && !strings.ContainsAny(basepath, ".well-known") && !strings.ContainsAny(basepath, ".htacess") {
		return true
	}

	if strings.ContainsAny(basepath, "node_modules") || strings.ContainsAny(basepath, "__py_cache__") {
		return true
	}

	return false
}

func Sync(client *sftp.Client, src string, dest string) error {
	err := client.MkdirAll(dest)
	if err != nil {
		return xerrors.Errorf("error creating %s directory %w", dest, err)
	}

	// Delete extra files
	walker := client.Walk(dest)
	for walker.Step() {
		localpath, err := filepath.Rel(dest, walker.Path())
		if err != nil {
			return err
		}

		remotepath := filepath.Join(dest, localpath)

		localfile, err := os.Stat(localpath)
		if err != nil {
			if os.IsNotExist(err) {
				// The file is present on remote but not locally
				if err := client.Remove(remotepath); err != nil {
					return xerrors.Errorf("error removing remote path %s: %w", remotepath, err)
				}
			}
		}

		remotefile, err := client.Stat(remotepath)
		if err != nil {
			return err
		}

		// Both are directories
		if localfile.IsDir() && remotefile.IsDir() {
			return nil
		}

		// Both are files
		if !localfile.IsDir() && !remotefile.IsDir() {
			if localfile.Size() == remotefile.Size() {
				identical, err := isIdentical(localpath, remotepath)
				if err != nil {
					return err
				}

				if identical {
					return nil
				}
			}
		}

		// files not identical
		// or one is a dir and the other is a file
		if err := client.Remove(remotepath); err != nil {
			return err
		}
	}

	// Create new files
	_fs := os.DirFS(src)
	fs.WalkDir(_fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if skipFile(path) {
			logging.Debugf("Path %s skipped", path)
			return nil
		}

		localpath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		remotepath := filepath.Join(dest, localpath)

		localfile, err := os.Stat(localpath)
		if err != nil {
			return err
		}

		if localfile.IsDir() {
			if err := client.MkdirAll(remotepath); err != nil {
				return err
			}
			return nil
		}

		return syncFile(client, localpath, remotepath)
	})

	return nil
}

func isIdentical(path1, path2 string) (bool, error) {
	buffer := make([]byte, 10*mega)
	h1 := md5.New()
	h2 := md5.New()

	f1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(path2)
	if err != nil {
		return false, err
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

func syncFile(client *sftp.Client, localpath, remotepath string) error {
	if remotef, err := client.Create(remotepath); err != nil {
		localf, err := os.Open(localpath)
		if err != nil {
			return err
		}

		_, err = io.Copy(remotef, localf)
		if err != nil {
			return err
		}
	}

	return nil
}
