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
	"github.com/spf13/cobra"
)

type ZpmRefreshCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRefreshCommand() *ZpmRefreshCommand {
	cmd := &ZpmRefreshCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "refresh"
	cmd.Short = "Refresh ZPM repository metadata"
	cmd.Long = "Refresh ZPM repository metadata"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRefreshCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRefreshCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Refresh()
	if err != nil {
		z.Fatal(err.Error())
	}
	return nil
}
