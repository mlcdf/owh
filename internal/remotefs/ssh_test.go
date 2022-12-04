package remotefs_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mlcdf.fr/owh/internal/remotefs"
	"go.mlcdf.fr/owh/internal/remotefs/sshtest"
)

func TestSync(t *testing.T) {
	container := sshtest.MustStart()

	defer container.Nuke()

	remote, err := sshtest.Connect()
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
			name:    "",
			src:     "",
			dest:    "yolo",
			wantErr: true,
			err:     remotefs.ErrEmptyStringSrc,
		},
		{
			name:    "",
			src:     "yolo",
			dest:    "",
			wantErr: true,
			err:     remotefs.ErrEmptyStringDest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := remote.Sync(test.src, test.dest)
			if !test.wantErr {
				require.NoError(t, err)
			}

			if test.err != nil {
				require.ErrorIs(t, test.err, err, "yolo")
			}

			output, err := remote.Run("tree " + test.dest)
			require.NoError(t, err)

			require.Equal(t, "", output)
		})
	}
}
