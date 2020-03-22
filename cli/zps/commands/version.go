/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package commands

import (
	"github.com/fezz-io/zps/cli"
	"github.com/spf13/cobra"
)

var Version string

type ZpsVersionCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsVersionCommand() *ZpsVersionCommand {
	cmd := &ZpsVersionCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "version"
	cmd.Short = "Show version"
	cmd.Long = "Show version"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsVersionCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsVersionCommand) run(cmd *cobra.Command, args []string) error {
	z.Out("ZPS v" + Version)
	return nil
}
