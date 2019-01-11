package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"runtime"
)

var (
	// Version is populated by govvv.
	Version = "untouched"
	// BuildDate is populated by govvv.
	BuildDate string
	// GitCommit is populated by govvv.
	GitCommit string
	// GitBranch is populated by govvv.
	GitBranch string
	// GitState is populated by govvv.
	GitState string
	// GitSummary is populated by govvv.
	GitSummary string
)

// VersionMsg is the message that is shown after by help command.
const VersionMsg = `version     : %s
build date  : %s
go version  : %s
go compiler : %s
platform    : %s/%s
git commit  : %s
git branch  : %s
git state   : %s
git summary : %s
`

// versionCmd contains the options for the version command.
type versionCmd struct{}

// Run prints the app version.
func (s *versionCmd) Run(ctx *kong.Context) error {
	fmt.Printf(VersionMsg, Version, BuildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH,
		GitCommit, GitBranch, GitState, GitSummary)
	return nil
}
