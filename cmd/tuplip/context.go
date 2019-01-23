package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/gofunky/automi/collectors"
	"github.com/gofunky/automi/stream"
	"github.com/gofunky/tuplip/pkg/tupliplib"
	"github.com/oleiade/reflections"
	"os"
	"strings"
)

// rootCmd wraps the root command as runnable.
type rootCmd interface {
	// Run executes the command given a tuplip source.
	run(src *tupliplib.TuplipSource) (*stream.Stream, error)
}

// tuplipContext provides the options and the interface to the tupliplib.
type tuplipContext struct {
	tupliplib.Tuplip `embed:""`
}

// sourceOption defines a command branch to determine the source of the tag vectors.
type sourceOption struct {
	stdinOption `embed:""`
	fileOption  `embed:""`
	paramOption `embed:""`
}

// fromRepositoryOption defines a command branch to determine the source of the tag vectors and a repository name.
type fromRepositoryOption struct {
	// Repository opens a positional argument in the command.
	Repository struct {
		// From command determines the source of the tag vectors that need the repository.
		From struct {
			sourceOption `embed:""`
		} `cmd:"" help:"determine the source of the tag vectors"`
		// Repository is the Docker Hub repository of the root tag vector in the format `organization/repository`.
		Repository string `arg:"" env:"DOCKER_REPOSITORY" help:"the Docker Hub repository of the root tag vector in the format 'organization/repository'"`
	} `arg:""`
}

// sourceTagOption defines a command branch to determine the source of the tag vectors, repository name, and source tag.
type sourceTagOption struct {
	// CheckSemver flag enables semantic version checks
	CheckSemver bool `short:"c" help:"check versioned tag vectors for valid semantic version syntax"`
	// SourceTag opens a positional argument in the command.
	SourceTag struct {
		// From command determines the source of the tag vectors.
		From sourceOption `cmd:"" help:"determine the source of the tag vectors"`
		// SourceTag is the tag of the source image that is to be tagged.
		SourceTag string `arg:"" help:"the source tag of the image that should receive the new tags (skippable)"`
		// To command defines the target repository.
		To struct {
			fromRepositoryOption `embed:""`
		} `cmd:"" help:"set a target repository (skippable)"`
	} `arg:""`
}

// processingFlags define flags for tag processing.
type processingFlags struct {
	// Straight defines a flag for straight processing without a build.
	Straight bool `short:"s" help:"use the input tags directly without any mixing"`
}

// toRoot determines the root command and passes the given tuplip source to it.
func (t tuplipContext) toRoot(ctx *kong.Context, src *tupliplib.TuplipSource) error {
	command := strings.SplitN(ctx.Command(), " ", 2)[0]
	capCommand := strings.Title(command)
	cmd, err := reflections.GetField(cli, capCommand)
	if err != nil {
		return err
	}
	rootCmd := cmd.(rootCmd)
	stm, err := rootCmd.run(src)
	if err != nil {
		return err
	}
	return t.write(stm)
}

// write the results from the given tuplip stream to the stdout.
func (t tuplipContext) write(stream *stream.Stream) error {
	lineSplit := func(input string) string {
		return fmt.Sprintln(input)
	}
	stream.Map(lineSplit)
	writer := collectors.Writer(bufio.NewWriter(os.Stdout))
	stream.Into(writer)
	if err := <-stream.Open(); err != nil {
		return err
	}
	return nil
}
