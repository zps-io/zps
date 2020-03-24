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

type ZpsConfigureCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsConfigureCommand() *ZpsConfigureCommand {
	cmd := &ZpsConfigureCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "configure [PKG ...]"
	cmd.Short = "Configure packages"
	cmd.Long = "Configure packages"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("profile", "default", "Profile for configuration context")

	return cmd
}

func (z *ZpsConfigureCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsConfigureCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	profile, _ := cmd.Flags().GetString("profile")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Configure(cmd.Flags().Args(), profile)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
