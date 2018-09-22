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
	"github.com/solvent-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpkgRootCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgRootCommand() *ZpkgRootCommand {
	cmd := &ZpkgRootCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zpkg"
	cmd.Short = "ZPKG work with ZPKG files"
	cmd.Long = "ZPKG work with ZPKG files"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.PersistentFlags().Bool("no-color", false, "Disable color")

	cmd.AddCommand(NewZpkgBuildCommand().Command)
	cmd.AddCommand(NewZpkgContentsCommand().Command)
	cmd.AddCommand(NewZpkgExtractCommand().Command)
	cmd.AddCommand(NewZpkgInfoCommand().Command)
	cmd.AddCommand(NewZpkgManifestCommand().Command)

	return cmd
}

func (z *ZpkgRootCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgRootCommand) run(cmd *cobra.Command, args []string) error {
	z.Help()
	return nil
}
