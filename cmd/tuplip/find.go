package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// findCmd contains the options for the find command.
type findCmd struct {
	// In command defines the target repository.
	In struct {
		fromRepositoryOption `embed:""`
	} `cmd:"" help:"set a target repository (skippable)"`
	// From command determines the source of the tag vectors.
	From fileOption `cmd:"" help:"determine the source of the tag vectors"`
}

// run implements main.rootCmd.run by executing the find.
func (s findCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	if r := s.In.Repository.Repository; r != "" {
		src.Repository = r
	}
	return src.Find()
}
