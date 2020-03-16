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

	"github.com/fezz-io/zps/cli/zps/commands"
)

func main() {
	command := commands.NewZpsRootCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
