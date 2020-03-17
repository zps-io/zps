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

type ZpsRootCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRootCommand() *ZpsRootCommand {
	cmd := &ZpsRootCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zps"
	cmd.Short = "ZPS the Z Package System"
	cmd.Long = "ZPS the Z Package System"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.PersistentFlags().Bool("no-color", false, "Disable color")
	cmd.PersistentFlags().String("image", "", "ZPS image name/id")

	cmd.AddCommand(NewZpsCacheCommand().Command)
	cmd.AddCommand(NewZpsChannelCommand().Command)
	cmd.AddCommand(NewZpsContentsCommand().Command)
	cmd.AddCommand(NewZpsFetchCommand().Command)
	cmd.AddCommand(NewZpsFreezeCommand().Command)
	cmd.AddCommand(NewZpsImageCommand().Command)
	cmd.AddCommand(NewZpsInfoCommand().Command)
	cmd.AddCommand(NewZpsInstallCommand().Command)
	cmd.AddCommand(NewZpsListCommand().Command)
	cmd.AddCommand(NewZpsPkiCommand().Command)
	cmd.AddCommand(NewZpsPlanCommand().Command)
	cmd.AddCommand(NewZpsPublishCommand().Command)
	cmd.AddCommand(NewZpsRefreshCommand().Command)
	cmd.AddCommand(NewZpsRemoveCommand().Command)
	cmd.AddCommand(NewZpsRepoCommand().Command)
	cmd.AddCommand(NewZpsStatusCommand().Command)
	cmd.AddCommand(NewZpsThawCommand().Command)
	cmd.AddCommand(NewZpsTransactionCommand().Command)
	cmd.AddCommand(NewZpsUpdateCommand().Command)
	cmd.AddCommand(NewZpsVersionCommand().Command)
	cmd.AddCommand(NewZpsZpkgCommand().Command)

	return cmd
}

func (z *ZpsRootCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRootCommand) run(cmd *cobra.Command, args []string) error {
	return z.Help()
}
