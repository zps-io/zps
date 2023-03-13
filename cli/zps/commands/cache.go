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
	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
)

type ZpsCacheCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsCacheCommand() *ZpsCacheCommand {
	cmd := &ZpsCacheCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "cache"
	cmd.Short = "Manage metadata and file cache"
	cmd.Long = "Manage metadata and file cache"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsCacheCleanCommand().Command)
	cmd.AddCommand(NewZpsCacheClearCommand().Command)
	return cmd
}

func (z *ZpsCacheCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsCacheCommand) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
