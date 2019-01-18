package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// pushCmd contains the options for the push command.
type pushCmd struct {
	sourceTagOption `embed:""`
	processingFlags `embed:""`
	// To command defines the target repository.
	To struct {
		fromRepositoryOption `embed:""`
	} `cmd:"" help:"set a target repository"`
	// From command determines the source of the tag vectors.
	From sourceOption `cmd:"" help:"determine the source of the tag vectors"`
}

// run implements main.rootCmd.run by executing the tagging, and then the pushing process.
func (s pushCmd) run(src *tupliplib.TuplipSource) (stream *stream.Stream, err error) {
	repo := s.sourceTagOption.SourceTag.To.Repository.Repository + s.To.Repository.Repository
	if repo != "" {
		src.Repository = repo
	}
	if !s.processingFlags.Straight {
		stream = src.Build(s.CheckSemver)
	}
	if sourceTag := s.sourceTagOption.SourceTag.SourceTag; sourceTag != "" {
		if stream, err = src.Tag(sourceTag); err != nil {
			return
		}
	}
	stream, err = src.Push()
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
