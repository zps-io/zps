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

	"github.com/fezz-io/zps/cli"

	"github.com/fezz-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpsInstallCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsInstallCommand() *ZpsInstallCommand {
	cmd := &ZpsInstallCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "install [PKG ...]"
	cmd.Short = "Install packages"
	cmd.Long = "Install packages"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsInstallCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsInstallCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to install")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Install(cmd.Flags().Args())
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
