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
	"github.com/zps-io/zps/zpm"
)

type ZpsCacheClearCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsCacheClearCommand() *ZpsCacheClearCommand {
	cmd := &ZpsCacheClearCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "clear"
	cmd.Short = "Clear ZPS image cache"
	cmd.Long = "Clear ZPS image cache"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsCacheClearCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsCacheClearCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.CacheClear()
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
