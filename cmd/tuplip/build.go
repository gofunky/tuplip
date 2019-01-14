package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// buildCmd contains the options for the build command.
type buildCmd struct {
	// CheckSemver flag enables semantic version checks
	CheckSemver bool `short:"c" help:"check versioned tag vectors for valid semantic version syntax"`
	// From command determines the source of the tag vectors.
	From sourceOption `cmd:"" help:"determine the source of the tag vectors"`
}

// run implements main.rootCmd.run by executing the build.
func (s buildCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	return src.Build(s.CheckSemver), nil
}
