package main

import (
	"os"

	"github.com/solvent-io/zps/cli/zpkg/commands"
)

func main() {
	command := commands.NewZpkgRootCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
