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

type ZpsRepoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRepoCommand() *ZpsRepoCommand {
	cmd := &ZpsRepoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "repo"
	cmd.Short = "Manage repositories"
	cmd.Long = "Manage repositories"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsRepoInitCommand().Command)
	cmd.AddCommand(NewZpsRepoContentsCommand().Command)
	cmd.AddCommand(NewZpsRepoListCommand().Command)
	cmd.AddCommand(NewZpsRepoUpdateCommand().Command)
	cmd.AddCommand(NewZpsRepoUnlockCommand().Command)
	return cmd
}

func (z *ZpsRepoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRepoCommand) run(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
