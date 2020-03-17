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

type ZpsPkiKeyPairCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiKeyPairCommand() *ZpsPkiKeyPairCommand {
	cmd := &ZpsPkiKeyPairCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "keypair"
	cmd.Short = "Manage ZPS pki signing key pairs"
	cmd.Long = "Manage ZPS pki signing key pairs"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsPkiKeyPairImportCommand().Command)
	cmd.AddCommand(NewZpsPkiKeyPairListCommand().Command)
	cmd.AddCommand(NewZpsPkiKeyPairRemoveCommand().Command)

	return cmd
}

func (z *ZpsPkiKeyPairCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiKeyPairCommand) run(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
