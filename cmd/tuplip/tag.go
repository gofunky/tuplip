package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// tagCmd contains the options for the tag command.
type tagCmd struct {
	sourceTagOption `embed:""`
	processingFlags `embed:""`
}

// run implements main.rootCmd.run by executing the tagging process.
func (s tagCmd) run(src *tupliplib.TuplipSource) (stream *stream.Stream, err error) {
	if r := s.sourceTagOption.SourceTag.To.Repository.Repository; r != "" {
		src.Repository = r
	}
	if !s.processingFlags.Straight {
		stream = src.Build(s.CheckSemver)
	}
	stream, err = src.Tag(s.sourceTagOption.SourceTag.SourceTag)
	if err != nil {
		return nil, err
	}
	if !cli.Verbose {
		stream.Filter(func(in string) bool {
			return false
		})
	}
	return
}
