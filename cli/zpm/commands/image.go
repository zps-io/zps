/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package commands

import (
	"github.com/fezz-io/zps/cli"

	"github.com/spf13/cobra"
)

type ZpmImageCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmImageCommand() *ZpmImageCommand {
	cmd := &ZpmImageCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "image"
	cmd.Short = "Manage ZPS images"
	cmd.Long = "Manage ZPS images"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpmImageInitCommand().Command)
	cmd.AddCommand(NewZpmImageCurrentCommand().Command)
	cmd.AddCommand(NewZpmImageListCommand().Command)
	return cmd
}

func (z *ZpmImageCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmImageCommand) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
