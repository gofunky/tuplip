package main

import (
	"github.com/mitchellh/cli"
	"log"
	"os"
)

// GitVersion is to be replaced by the build version during build time.
var GitVersion = "dev"

// main builds a command factory and starts it for the binary.
func main() {
	c := cli.NewCLI("tuplip", GitVersion)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"exec": func() (command cli.Command, e error) {
			return new(ExecCommand), nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)

}
