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

type ZpmZpkgCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmZpkgCommand() *ZpmZpkgCommand {
	cmd := &ZpmZpkgCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zpkg"
	cmd.Short = "Work with ZPKG files"
	cmd.Long = "Work with ZPKG files"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpmZpkgBuildCommand().Command)
	cmd.AddCommand(NewZpmZpkgContentsCommand().Command)
	cmd.AddCommand(NewZpmZpkgExtractCommand().Command)
	cmd.AddCommand(NewZpmZpkgInfoCommand().Command)
	cmd.AddCommand(NewZpmZpkgManifestCommand().Command)
	cmd.AddCommand(NewZpmZpkgSignCommand().Command)
	cmd.AddCommand(NewZpmZpkgVerifyCommand().Command)

	return cmd
}

func (z *ZpmZpkgCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmZpkgCommand) run(cmd *cobra.Command, args []string) error {
	z.Help()
	return nil
}
