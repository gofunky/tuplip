package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// findCmd contains the options for the find command.
type findCmd struct {
	fromRepositoryOption `embed`
}

// run implements main.rootCmd.run by executing the find.
func (s findCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	if r := s.Repository.Repository; r != "" {
		src.Repository = r
	}
	return src.Find()
}
