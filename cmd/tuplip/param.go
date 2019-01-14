package main

import (
	"github.com/alecthomas/kong"
)

// paramOption defines a command branch that contains only the param command.
type paramOption struct {
	// Param to pass the tag vectors as arguments.
	Param paramCmd `arg:"" help:"pass the tag vectors as arguments"`
}

// paramCmd defines a command to pass the tag vectors as parameter.
type paramCmd struct {
	Context tuplipContext `embed:""`
	// Param is the parameter that contains the tag vectors.
	Param []string `arg:"" help:"the space-separated list of tag vectors"`
}

// Run implements a dynamic interface from kong by executing a command using given param argument as input.
func (c paramCmd) Run(ctx *kong.Context) error {
	tuplip := c.Context.Tuplip
	src := (&tuplip).FromSlice(c.Param)
	return c.Context.toRoot(ctx, src)
}
