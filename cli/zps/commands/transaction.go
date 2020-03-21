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

type ZpsTransactionCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsTransactionCommand() *ZpsTransactionCommand {
	cmd := &ZpsTransactionCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "transaction"
	cmd.Short = "Manage ZPS image transactions"
	cmd.Long = "Manage ZPS image transactions"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpsTransactionListCommand().Command)
	return cmd
}

func (z *ZpsTransactionCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsTransactionCommand) run(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
