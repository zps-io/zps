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
	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
)

type ZpsZpkgCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsZpkgCommand() *ZpsZpkgCommand {
	cmd := &ZpsZpkgCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zpkg"
	cmd.Short = "Manage ZPKG files"
	cmd.Long = "Manage ZPKG files"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsZpkgBuildCommand().Command)
	cmd.AddCommand(NewZpsZpkgContentsCommand().Command)
	cmd.AddCommand(NewZpsZpkgExtractCommand().Command)
	cmd.AddCommand(NewZpsZpkgInfoCommand().Command)
	cmd.AddCommand(NewZpsZpkgManifestCommand().Command)
	cmd.AddCommand(NewZpsZpkgSignCommand().Command)
	cmd.AddCommand(NewZpsZpkgValidateCommand().Command)

	return cmd
}

func (z *ZpsZpkgCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsZpkgCommand) run(cmd *cobra.Command, args []string) error {
	z.Help()
	return nil
}
