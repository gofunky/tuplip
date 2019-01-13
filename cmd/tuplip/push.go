package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// pushCmd contains the options for the push command.
type pushCmd struct {
	sourceTagOption `embed`
}

// run implements main.rootCmd.run by executing the tagging, and then the pushing process.
func (s pushCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	if r := s.sourceTagOption.SourceTag.Repository.Repository; r != "" {
		src.Repository = r
	}
	return src.Push(s.CheckSemver, s.sourceTagOption.SourceTag.SourceTag)
}
