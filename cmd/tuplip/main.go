package main

import (
	"github.com/alecthomas/kong"
	"os"
)

// cli references the cli command objects.
var cli struct {
	Version versionCmd `cmd help:"display the app version"`
	Help    helpCmd    `cmd help:"show help for a command"`
	Build   buildCmd   `cmd help:"build all possible Docker tags from the given tag vectors"`
	Find    findCmd    `cmd help:"find the most appropriate Docker tag in the given repository"`
}

// main builds a command factory and starts it for the binary.
func main() {
	ctx := kong.Parse(&cli,
		kong.Writers(os.Stderr, os.Stderr),
		kong.Name("tuplip"),
		kong.Description("tuplip is a convention-forming Docker tag generator and checker."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.Vars{
			"version": Version,
		},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
