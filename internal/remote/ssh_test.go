package remote_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.mlcdf.fr/owh/internal/remote"
	"go.mlcdf.fr/owh/tests/sshtest"

	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	t.Parallel()
	container := sshtest.MustStart(t)

	defer container.Nuke(t)

	remotefs, err := remote.Connect(
		&remote.Config{Host: "localhost", Port: sshtest.Port, SSHConfig: sshtest.SSHKeyConfig()},
	)
	require.NoError(t, err)

	tests := []struct {
		name       string
		src        string
		dest       string
		wantRemote string
		wantErr    bool
		err        error
	}{
		{
			name:    "empty src",
			src:     "",
			dest:    "yolo",
			wantErr: true,
			err:     remote.ErrEmptyStringSrc,
		},
		{
			name:    "empty dest",
			src:     "yolo",
			dest:    "",
			wantErr: true,
			err:     remote.ErrEmptyStringDest,
		},
		{
			name: "sync www",
			src:  "fixtures/www",
			dest: "www",
		},
		{
			name: "sync with-subdir",
			src:  "fixtures/with-subdir",
			dest: "with-subdir",
		},
		{
			name: "with-subdir after some changes",
			src:  "fixtures/with-subdir-v2",
			dest: "with-subdir",
		},
		{
			name: "sync back with-subdir",
			src:  "fixtures/with-subdir",
			dest: "with-subdir",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := remotefs.Sync(test.src, test.dest)
			if !test.wantErr {
				require.NoError(t, err)
			}

			if test.err != nil {
				require.ErrorIs(t, test.err, err)
				return
			}

			require.Equal(t,
				localtree(t, test.src),
				remotetree(t, remotefs, test.dest),
			)
		})
	}
}

func localtree(t *testing.T, arg string) string {
	t.Helper()

	wd := filepath.Dir(arg)
	dir, err := filepath.Rel(wd, arg)
	require.NoError(t, err)

	cmd := exec.Command("tree", dir)
	cmd.Dir = wd

	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	// clean output

	toReplace := map[string]string{
		"\u00a0": " ",
		"└":      "`",
		"─":      "-",
		"├":      "|",
		"│":      "|",
	}

	want := string(output)

	for _from, to := range toReplace {
		want = strings.ReplaceAll(want, _from, to)
	}

	return removeFirstLine(want)
}

func remotetree(t *testing.T, remotefs *remote.Client, arg string) string {
	t.Helper()

	output, err := remotefs.Run(fmt.Sprintf("tree %s", arg))
	t.Log(output)
	require.NoError(t, err)

	return removeFirstLine(output)
}

func removeFirstLine(txt string) string {
	lines := strings.Split(txt, "\n")

	out := []string{}

	for index, line := range lines {
		if index != 0 {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n")
}
