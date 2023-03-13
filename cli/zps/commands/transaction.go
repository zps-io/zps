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

type ZpsTransactionCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsTransactionCommand() *ZpsTransactionCommand {
	cmd := &ZpsTransactionCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "transaction"
	cmd.Short = "Manage transactions"
	cmd.Long = "Manage transactions"
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
