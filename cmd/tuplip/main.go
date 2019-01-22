package main

import (
	"github.com/alecthomas/kong"
	"github.com/francoispqt/onelog"
	"github.com/gofunky/tuplip/pkg/help"
	"github.com/gofunky/tuplip/pkg/tupliplib"
	"os"
)

// cli references the cli command objects.
var cli struct {
	// Version describes the version command.
	Version versionCmd `cmd:"" help:"display the app version"`
	// Help describes the help command.
	Help helpCmd `cmd:"" help:"show help for a command"`
	// Graph prints the command graph.
	Graph graphCmd `cmd:"" help:"print the command graph"`
	// Build describes the build command.
	Build buildCmd `cmd:"" help:"build Docker tags from the given tag vectors"`
	// Tag describes the tag command.
	Tag tagCmd `cmd:"" help:"tag the given source image with the Docker tags from the given tag vectors"`
	// Push describes the push command.
	Push pushCmd `cmd:"" help:"tag and push the given source image with the Docker tags from the given tag vectors"`
	// Find describes the find command.
	Find findCmd `cmd:"" help:"find the most appropriate Docker tag in the given repository"`
	// Verbose mode enables detailed logging messages.
	Verbose bool `short:"v" help:"print detailed logging messages"`
}

// main builds a command factory and starts it for the binary.
func main() {
	ctx := kong.Parse(&cli,
		kong.Writers(os.Stderr, os.Stderr),
		kong.Name("tuplip"),
		kong.Description("tuplip is a convention-forming Docker tag generator and checker."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree:     true,
			Indenter: kong.TreeIndenter,
		}),
		kong.Vars{
			"version": Version,
		},
	)
	var levels = uint8(onelog.WARN | onelog.ERROR | onelog.FATAL)
	if cli.Verbose {
		levels = uint8(onelog.INFO | levels)
	}
	logger := onelog.New(os.Stderr, levels)
	tupliplib.UseLogger(logger)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
