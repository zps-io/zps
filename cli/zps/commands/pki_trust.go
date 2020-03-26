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

type ZpsPkiTrustCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiTrustCommand() *ZpsPkiTrustCommand {
	cmd := &ZpsPkiTrustCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "trust"
	cmd.Short = "Manage ZPS pki trust store"
	cmd.Long = "Manage ZPS pki trust store"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsPkiTrustFetchCommand().Command)
	cmd.AddCommand(NewZpsPkiTrustImportCommand().Command)
	cmd.AddCommand(NewZpsPkiTrustListCommand().Command)
	cmd.AddCommand(NewZpsPkiTrustRemoveCommand().Command)

	return cmd
}

func (z *ZpsPkiTrustCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiTrustCommand) run(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
