package main

import (
	"github.com/alecthomas/kong"
)

// fileOption defines a command branch that contains only the file command.
type fileOption struct {
	// File to read the tag vectors from a Dockerfile.
	File fileCmd `cmd:"" help:"read the tag vectors from a Dockerfile"`
}

// fileCmd defines a command to read tag vectors from a Dockerfile.
type fileCmd struct {
	Context tuplipContext `embed:""`
	// File is the Dockerfile that contains the vectors as FROM instructions.
	File string `arg:"" type:"existingfile" help:"the Dockerfile containing the vectors as FROM instructions"`
	// RootVersion overrides the version of the root tag vector from the Dockerfile.
	RootVersion string `short:"r" help:"override the version of the root tag vector from the Dockerfile"`
}

// Run implements a dynamic interface from kong by executing a command using given file argument as input.
func (c fileCmd) Run(ctx *kong.Context) error {
	tuplip := c.Context.Tuplip
	if src, err := (&tuplip).FromFile(c.File, c.RootVersion); err != nil {
		return err
	} else {
		return c.Context.toRoot(ctx, src)
	}
}
