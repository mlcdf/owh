//go:build !testbincover
// +build !testbincover

package tests

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/confluentinc/bincover"
	"github.com/stretchr/testify/require"
	"go.mlcdf.fr/owh/internal/config"
)

const binPath = "../dist/owh.test"

var collector *bincover.CoverageCollector

func TestMain(m *testing.M) {
	code := func() int {
		output, err := exec.Command("../scripts/build-test-binary.sh").CombinedOutput()
		if err != nil {
			log.Printf("err: %s, output: %s", err, output)
			return 1
		}

		collector = bincover.NewCoverageCollector("../dist/coverage.out", true)

		err = collector.Setup()
		if err != nil {
			log.Printf("err: %s", err)
			return 1
		}

		code := m.Run()

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recovered from ", r)
			}
		}()

		err = collector.TearDown()
		if err != nil {
			log.Printf("err: %s", err)
			return 1
		}

		err = os.Remove(binPath)
		if err != nil {
			log.Printf("err: %s", err)
			return 1
		}

		return code
	}()

	os.Exit(code)
}

// cleanEnv unset existing env vars
func cleanEnv(t *testing.T) {
	t.Helper()

	for _, e := range os.Environ() {
		key := strings.SplitN(e, "=", 1)[0]
		if strings.HasPrefix(key, config.ENV_PREFIX) {
			err := os.Setenv(key, "")
			require.NoError(t, err)
		}
	}
}
