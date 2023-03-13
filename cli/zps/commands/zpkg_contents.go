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

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
)

type ZpsZpkgContentsCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsZpkgContentsCommand() *ZpsZpkgContentsCommand {
	cmd := &ZpsZpkgContentsCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "contents [ZPKG_PATH]"
	cmd.Short = "List contents of a ZPKG"
	cmd.Long = "List contents of a ZPKG"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsZpkgContentsCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsZpkgContentsCommand) run(cmd *cobra.Command, args []string) error {
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

	output, err := mgr.ZpkgContents(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(columnize.SimpleFormat(output))

	return err
}
