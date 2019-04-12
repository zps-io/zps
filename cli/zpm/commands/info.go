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
	"github.com/ryanuber/columnize"
	"github.com/fezz-io/zps/cli"

	"github.com/fezz-io/zps/zpm"

	"errors"

	"github.com/spf13/cobra"
)

type ZpmInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmInfoCommand() *ZpmInfoCommand {
	cmd := &ZpmInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info"
	cmd.Short = "Show installed package metadata"
	cmd.Long = "Show installed package metadata"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmInfoCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	var err error

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide a package uri to query")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	info, err := mgr.Info(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	if info != nil {
		z.Out(columnize.SimpleFormat(z.Colorize(info)))
	}

	return nil
}
