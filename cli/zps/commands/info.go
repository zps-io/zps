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
	"github.com/zps-io/zps/cli"

	"github.com/zps-io/zps/zpm"

	"errors"

	"github.com/spf13/cobra"
)

type ZpsInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsInfoCommand() *ZpsInfoCommand {
	cmd := &ZpsInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info [PKG]"
	cmd.Short = "Show installed package metadata"
	cmd.Long = "Show installed package metadata"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsInfoCommand) run(cmd *cobra.Command, args []string) error {
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
