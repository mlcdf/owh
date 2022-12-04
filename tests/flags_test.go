package tests

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantOutput    string
		outputPattern *regexp.Regexp
		wantExitCode  int
	}{
		{
			name:          "print version with short flag",
			args:          []string{"-v"},
			outputPattern: regexp.MustCompile("  Version         = .*"),
			wantExitCode:  0,
		},
		{
			name:          "print version with full flag",
			args:          []string{"--version"},
			outputPattern: regexp.MustCompile("  Version         = .*"),
			wantExitCode:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output, exitCode, err := collector.RunBinary(binPath, "TestBincoverRunMain", []string{}, tt.args, "")
			require.NoError(t, err)

			if tt.outputPattern != nil {
				require.Regexp(t, tt.outputPattern, output)
			} else {
				require.Equal(t, tt.wantOutput, output)
			}
			require.Equal(t, tt.wantExitCode, exitCode)
		})
	}
}

func TestRequiredFlags(t *testing.T) {
	cleanEnv(t)

	tests := []struct {
		name          string
		args          string
		env           []string
		wantOutput    string
		outputPattern *regexp.Regexp
		wantExitCode  int
	}{
		{
			name:         "whoami",
			args:         "whoami",
			wantOutput:   "Maxime Le Conte des Floris (lm207219-ovh)\n",
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, exitCode, err := collector.RunBinary(binPath, "TestBincoverRunMain", tt.env, strings.Split(tt.args, " "), "")
			require.NoError(t, err)

			if tt.outputPattern != nil {
				require.Regexp(t, tt.outputPattern, output)
			} else {
				require.Equal(t, tt.wantOutput, output)
			}
			require.Equal(t, tt.wantExitCode, exitCode)
		})
	}
}
