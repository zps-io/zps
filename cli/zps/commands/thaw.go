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

	"github.com/zps-io/zps/cli"

	"github.com/spf13/cobra"
	"github.com/zps-io/zps/zpm"
)

type ZpsThawCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsThawCommand() *ZpsThawCommand {
	cmd := &ZpsThawCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "thaw [PKG ...]"
	cmd.Short = "Un-freeze package version"
	cmd.Long = "Un-freeze package version"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsThawCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsThawCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to freeze")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Thaw(args)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
