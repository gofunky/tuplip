package main

import (
	"bufio"
	"fmt"
	"github.com/gofunky/automi/collectors"
	"github.com/gofunky/tuplip/pkg/tupliplib"
	"os"
	"strings"
)

// The execution command of the binary.
type ExecCommand struct{}

// Run starts the execution process of the tuplip lib.
func (c *ExecCommand) Run(args []string) int {
	tuplip := tupliplib.Tuplip{}
	for _, arg := range args {
		lowerArg := strings.ToLower(arg)
		if strings.Contains(lowerArg, "excludemajor") {
			tuplip.ExcludeMajor = true
		}
		if strings.Contains(lowerArg, "excludeminor") {
			tuplip.ExcludeMinor = true
		}
		if strings.Contains(lowerArg, "excludebase") {
			tuplip.ExcludeBase = true
		}
	}
	return execute(tuplip)
}

// Help gives the binary documentation that is printed in the help.
func (c *ExecCommand) Help() string {
	return "Create a power set of possible tag and version combinations and parse it in the Docker tagging style.\n" +
		"Each input tag from the pipeline will be used to create combinations.\n" +
		"For instance, 'latest foo' creates 'latest', 'foo', and 'latest-foo' to depict all possibilities.\n" +
		"\nFurthermore, tuplip also accepts version suffixes (e.g., as commonly used for alpine image tag suffixes).\n" +
		"To let tuplip create all possible images for the semantic version of the dependency, " +
		"pass the tag alias colon-separated from the version.\n" +
		"For example, use 'alpine:3.8' to create the images 'alpine', 'alpine3', and 'alpine3.8' that are then further " +
		"combined with your other given input tags.\n" +
		"\nBeside the versions of the image dependencies, use the version product for your image version " +
		"by setting '_' as the version alias (e.g., as in '_:1.0.0').\n" +
		"\nOptionally, use the following arguments to reduce the number of returned tags:\n" +
		"--excludeMajor to exclude the major versions from the power set\n" +
		"--excludeMinor to exclude the minor versions from the power set\n" +
		"--excludeBase to exclude the base without version suffix from the power set\n"
}

// Synopsis gives a short description of the purpose.
func (c *ExecCommand) Synopsis() string {
	return "generate all Docker tags"
}

// execute reads the input and prints the output of tuplip.
func execute(tuplip tupliplib.Tuplip) int {
	reader := bufio.NewReader(os.Stdin)
	tuplipStream := tuplip.FromReader(reader)
	lineSplit := func(input string) string {
		return fmt.Sprintln(input)
	}
	tuplipStream.Map(lineSplit)
	writer := collectors.Writer(bufio.NewWriter(os.Stdout))
	tuplipStream.Into(writer)
	if err := <-tuplipStream.Open(); err != nil {
		_, printErr := fmt.Fprintf(os.Stderr, "tuplip failed due to an error: %v\n", err)
		if printErr != nil {
			panic(printErr)
		}
		return 1
	}
	return 0
}
