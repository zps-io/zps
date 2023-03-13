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
	"errors"

	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
	"github.com/zps-io/zps/zpm"
)

type ZpsRemoveCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRemoveCommand() *ZpsRemoveCommand {
	cmd := &ZpsRemoveCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "remove [PKG ...]"
	cmd.Short = "Remove packages"
	cmd.Long = "remove packages"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsRemoveCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRemoveCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to remove")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Remove(args)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
