package main

import (
	"os"

	"github.com/solvent-io/zps/cli/zpm/commands"
)

func main() {
	command := commands.NewZpmRootCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
