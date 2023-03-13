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
	"github.com/zps-io/zps/cli"

	"github.com/spf13/cobra"
)

type ZpsImageCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsImageCommand() *ZpsImageCommand {
	cmd := &ZpsImageCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "image"
	cmd.Short = "Manage images"
	cmd.Long = "Manage images"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsImageInitCommand().Command)
	cmd.AddCommand(NewZpsImageCurrentCommand().Command)
	cmd.AddCommand(NewZpsImageDeleteCommand().Command)
	cmd.AddCommand(NewZpsImageListCommand().Command)
	return cmd
}

func (z *ZpsImageCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsImageCommand) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
