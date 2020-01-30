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

	"github.com/fezz-io/zps/zpm"

	"github.com/fezz-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpmZpkgInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmZpkgInfoCommand() *ZpmZpkgInfoCommand {
	cmd := &ZpmZpkgInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info [ZPKG path]"
	cmd.Short = "Display ZPKG file information"
	cmd.Long = "Display ZPKG file information"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmZpkgInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmZpkgInfoCommand) run(cmd *cobra.Command, args []string) error {
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
