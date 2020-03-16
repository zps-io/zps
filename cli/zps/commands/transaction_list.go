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
	"github.com/fezz-io/zps/zpm"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type ZpsTransactionListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsTransactionListCommand() *ZpsTransactionListCommand {
	cmd := &ZpsTransactionListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List ZPS image transactions"
	cmd.Long = "List ZPS image transactions"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsTransactionListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsTransactionListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	transactions, err := mgr.TransActionList()
	if err != nil {
		z.Fatal(err.Error())
	}

	if transactions != nil {
		z.Info(columnize.SimpleFormat(transactions))
	}

	return nil
}
