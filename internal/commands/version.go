package commands

import (
	"fmt"
	"runtime"
	"strings"

	"go.mlcdf.fr/owh/internal/cmdutil"

	"go.mlcdf.fr/sally/build"
)

func Version(version string) {

	rows := []cmdutil.LabelValue{
		{Label: "Version", Value: build.Version},
		{Label: "Git Commit Date", Value: build.LastCommitTime.Format("2006-01-02")},
		{Label: "Git Commit Hash", Value: build.Revision},
		{Label: "Go Version", Value: strings.ReplaceAll(build.GoVersion, "go", "")},
		{Label: "Platform", Value: runtime.GOOS + "/" + runtime.GOARCH},
	}

	cmdutil.DescriptionTable("", rows)
	fmt.Println("\n  Written by Maxime Le Conte des Floris <git@mlcdf.fr>")
}
