package main

import (
	"fmt"
	"github.com/alecthomas/kong"
)

// HelpText contains the root cli description.
const HelpText = `
tuplip generates and checks Docker tags in a transparent and convention-forming way.

tuplip accepts three kinds of tag vectors (i.e., sub elements that define a tag) as the input,
namely (unversioned) alias vectors, (versioned) dependency vectors, and a (versioned) root vector.

The alias vector does not contain a version and is just passed as it. 
It can be used to identify a feature or a branch of the Docker image.

The dependency vector depicts the dependencies of the Docker image. All versioned sub layers, binaries, 
and imports of the image can be depicted with their regarding semantic version variants.
To let tuplip create all possible tuples for the semantic version of the dependency, 
pass the tag vector alias colon-separated from the version. For example, 
use 'alpine:3.8' to create the vector tuple 'alpine', 'alpine3',
and 'alpine3.8' that are then further combined with the other given input vectors.

Root vectors are similar to the dependency vectors but they don't contain the alias prefix. 
They are meant to be used for the core product that the Docker images represents. 
The root vector is always ordered before the remaining elements. 
A root vector is defined by setting '_' as the version alias (e.g., as in '_:1.0.0').

The tag vectors are fetched from the cli standard input by default, either line-separated or separated by the given
separator argument. Alternatively, in 'fromFile', a Dockerfile can be used instead of the standard input,
to parse its FROM instructions as vectors.

For more information, visit the GitHub project https://github.com/gofunky/tuplip.`

// helpCmd contains the options for the help command.
type helpCmd struct {
	Command []string `arg:"" optional:"" help:"the commands for which to show help" sep:" "`
}

// Run shows help.
func (h *helpCmd) Run(realCtx *kong.Context) error {
	ctx, err := kong.Trace(realCtx.Kong, h.Command)
	if err != nil {
		return err
	}
	if ctx.Error != nil {
		return ctx.Error
	}
	err = ctx.PrintUsage(false)
	if err != nil {
		return err
	}
	fmt.Println(HelpText)
	_, err = fmt.Fprintln(realCtx.Stdout)
	return err
}
