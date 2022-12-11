package remote_test

import (
	"fmt"
	"os/exec"
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
			name: "",
			src:  "fixtures/local/www",
			dest: "www",
		},
		{
			name: "",
			src:  "fixtures/local/with-subdir",
			dest: "with-subdir",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := remotefs.Run("ls")
			t.Log(output)
			require.NoError(t, err)

			err = remotefs.Sync(test.src, test.dest)
			if !test.wantErr {
				require.NoError(t, err)
			}

			if test.err != nil {
				require.ErrorIs(t, test.err, err)
				return
			}

			require.Equal(t,
				localtree(t, "fixtures/local", test.dest),
				remotetree(t, remotefs, test.dest),
			)
		})
	}
}

func localtree(t *testing.T, wd string, arg string) string {
	t.Helper()

	cmd := exec.Command("tree", arg)
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

	return want
}

func remotetree(t *testing.T, remotefs *remote.Client, arg string) string {
	t.Helper()

	output, err := remotefs.Run(fmt.Sprintf("tree %s", arg))
	t.Log(output)
	require.NoError(t, err)
	return output
}
