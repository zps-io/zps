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

	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpsZpkgSignCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsZpkgSignCommand() *ZpsZpkgSignCommand {
	cmd := &ZpsZpkgSignCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "sign [ZPKGFILE PATH]"
	cmd.Short = "Sign a ZPKG"
	cmd.Long = "Sign a ZPKG"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("work-path", "", "Work path for ZPKG creation")

	return cmd
}

func (z *ZpsZpkgSignCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsZpkgSignCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	workPath, _ := cmd.Flags().GetString("work-path")

	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.ZpkgSign(cmd.Flags().Arg(0), workPath)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
