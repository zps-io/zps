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

type ZpmRootCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRootCommand() *ZpmRootCommand {
	cmd := &ZpmRootCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zpm"
	cmd.Short = "ZPM is the package management component of ZPS the Z Package System"
	cmd.Long = "ZPM is the package management component of ZPS the Z Package System"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.PersistentFlags().Bool("no-color", false, "Disable color")
	cmd.PersistentFlags().String("image", "", "ZPS image name/id")

	cmd.AddCommand(NewZpmCacheCommand().Command)
	cmd.AddCommand(NewZpmChannelCommand().Command)
	cmd.AddCommand(NewZpmContentsCommand().Command)
	cmd.AddCommand(NewZpmFetchCommand().Command)
	cmd.AddCommand(NewZpmFreezeCommand().Command)
	cmd.AddCommand(NewZpmImageCommand().Command)
	cmd.AddCommand(NewZpmInfoCommand().Command)
	cmd.AddCommand(NewZpmInstallCommand().Command)
	cmd.AddCommand(NewZpmListCommand().Command)
	cmd.AddCommand(NewZpmPkiCommand().Command)
	cmd.AddCommand(NewZpmPlanCommand().Command)
	cmd.AddCommand(NewZpmPublishCommand().Command)
	cmd.AddCommand(NewZpmRefreshCommand().Command)
	cmd.AddCommand(NewZpmRemoveCommand().Command)
	cmd.AddCommand(NewZpmRepoCommand().Command)
	cmd.AddCommand(NewZpmStatusCommand().Command)
	cmd.AddCommand(NewZpmThawCommand().Command)
	cmd.AddCommand(NewZpmTransactionCommand().Command)
	cmd.AddCommand(NewZpmUpdateCommand().Command)
	cmd.AddCommand(NewZpmVersionCommand().Command)
	cmd.AddCommand(NewZpmZpkgCommand().Command)

	return cmd
}

func (z *ZpmRootCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRootCommand) run(cmd *cobra.Command, args []string) error {
	z.Help()
	return nil
}
