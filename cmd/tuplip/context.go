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
	tupliplib.Tuplip `embed`
}

// fileOption defines a command branch that contains only the file command.
type fileOption struct {
	// File to read the tag vectors from a Dockerfile.
	File fileCmd `cmd help:"read the tag vectors from a Dockerfile"`
}

// sourceOption defines a command branch to determine the source of the tag vectors.
type sourceOption struct {
	stdinOption `embed`
	fileOption  `embed`
	paramOption `embed`
}

// fileCmd defines a command to read tag vectors from a Dockerfile.
type fileCmd struct {
	// File opens a positional argument in the file command.
	File struct {
		// File is the Dockerfile that contains the vectors as FROM instructions.
		File string `arg type:"existingfile" help:"the Dockerfile containing the vectors as FROM instructions"`
	} `arg`
}

// toRoot determines the root command and passes the given tuplip source to it.
func (t tuplipContext) toRoot(ctx *kong.Context, src *tupliplib.TuplipSource) error {
	command := strings.SplitN(ctx.Command(), " ", 2)[0]
	capCommand := strings.Title(command)
	if cmd, err := reflections.GetField(cli, capCommand); err != nil {
		return err
	} else {
		rootCmd := cmd.(rootCmd)
		if stm, stmErr := rootCmd.run(src); stmErr != nil {
			return stmErr
		} else {
			return t.write(stm)
		}
	}
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
