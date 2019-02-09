/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package main

import (
	"os"
	"path"

	zpkgcmds "github.com/solvent-io/zps/cli/zpkg/commands"
	zpmcmds "github.com/solvent-io/zps/cli/zpm/commands"
)

func main() {
	switch path.Base(os.Args[0]) {
	case "zpkg":
		command := zpkgcmds.NewZpkgRootCommand()
		if err := command.Execute(); err != nil {
			os.Exit(1)
		}

	case "zpm":
		command := zpmcmds.NewZpmRootCommand()
		if err := command.Execute(); err != nil {
			os.Exit(1)
		}

	default:
		os.Exit(0)
	}
}
