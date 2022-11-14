// build allows access to common build info and tags
package build // import "go.mlcdf.fr/sally/build"

import (
	"fmt"
	"runtime/debug"
	"time"
)

var (
	// Version is set at build time using -ldflags="-X 'go.mlcdf.fr/sally/build.Version=v1.0.0'"
	Version        = "(devel)"
	Revision       = "unknown"
	LastCommitTime time.Time
	GoVersion      = "unknown"
)

func init() {
	buildInfo, ok := debug.ReadBuildInfo()

	if ok {
		GoVersion = buildInfo.GoVersion

		for _, kv := range buildInfo.Settings {
			switch kv.Key {
			case "vcs.revision":
				Revision = kv.Value
			case "vcs.time":
				LastCommitTime, _ = time.Parse(time.RFC3339, kv.Value)
			}
		}
	}
}

func String() string {
	return fmt.Sprintf("version %s (%s)", Version, LastCommitTime.Format("2006-01-02"))
}
