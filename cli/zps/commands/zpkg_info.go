/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package commands

import (
	"errors"

	"github.com/zps-io/zps/zpm"

	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
)

type ZpsZpkgInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsZpkgInfoCommand() *ZpsZpkgInfoCommand {
	cmd := &ZpsZpkgInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info [ZPKG_PATH]"
	cmd.Short = "Display ZPKG file information"
	cmd.Long = "Display ZPKG file information"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsZpkgInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsZpkgInfoCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	output, err := mgr.ZpkgInfo(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(output)

	return err
}
