package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// buildCmd contains the options for the build command.
type buildCmd struct {
	// CheckSemver flag enables semantic version checks
	CheckSemver bool `short:"c" help:"check versioned tag vectors for valid semantic version syntax"`
	// To command defines the target repository.
	To struct {
		fromRepositoryOption `embed:""`
	} `cmd:"" help:"set a target repository to be prefixed in the output set (skippable)"`
	// From command determines the source of the tag vectors.
	From sourceOption `cmd:"" help:"determine the source of the tag vectors"`
}

// run implements main.rootCmd.run by executing the build.
func (s buildCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	if r := s.To.Repository.Repository; r != "" {
		src.Repository = r
	}
	return src.Build(s.CheckSemver), nil
}
