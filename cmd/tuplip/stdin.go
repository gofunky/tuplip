package main

import (
	"bufio"
	"github.com/alecthomas/kong"
	"os"
)

// stdinOption defines a command branch that contains only the stdin command.
type stdinOption struct {
	// Stdin is used to read the tag vectors.
	Stdin stdinCmd `cmd:"" help:"read the tag vectors from the standard input"`
}

// stdinCmd defines a command to read tag vectors from a standard input.
type stdinCmd struct {
	Context tuplipContext `embed:""`
	// Separator that splits the separate tag vectors. The default separator is a single space.
	Separator string `optional:"" short:"s" env:"IFS" default:" " help:"the separator that splits the separate tag vectors from the stdin"`
}

// Run implements a dynamic interface from kong by executing a command using the stdin as input.
func (c stdinCmd) Run(ctx *kong.Context) error {
	reader := bufio.NewReader(os.Stdin)
	tuplip := c.Context.Tuplip
	src := (&tuplip).FromReader(reader, c.Separator)
	return c.Context.toRoot(ctx, src)
}
