package main

import (
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
)

// findCmd contains the options for the find command.
type findCmd struct {
	// From command determines the source of the tag vectors.
	From fileOption `cmd help:"determine the source of the tag vectors"`
	// Repository opens a positional argument in the find command.
	Repository struct {
		// From command determines the source of the tag vectors that need the repository.
		From struct {
			stdinOption `embed`
			paramOption `embed`
		} `cmd help:"determine the source of the tag vectors"`
		// Repository is the Docker Hub repository of the root tag vector in the format `organization/repository`.
		Repository string `arg env:"DOCKER_REPOSITORY" help:"the Docker Hub repository of the root tag vector in the format 'organization/repository'"`
	} `arg`
}

// run implements main.rootCmd.run by executing the find.
func (s findCmd) run(src *tupliplib.TuplipSource) (*stream.Stream, error) {
	if r := s.Repository.Repository; r != "" {
		src.Repository = r
	}
	return src.Find()
}
