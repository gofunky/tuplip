package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// tagCmd contains the options for the tag command.
type tagCmd struct {
	sourceTagOption `embed`
}

// run implements main.rootCmd.run by executing the tagging process.
func (s tagCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	if r := s.sourceTagOption.SourceTag.Repository.Repository; r != "" {
		src.Repository = r
	}
	return src.Tag(s.CheckSemver, s.sourceTagOption.SourceTag.SourceTag)
}
