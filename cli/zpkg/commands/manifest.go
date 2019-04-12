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
	"github.com/fezz-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpkgManifestCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgManifestCommand() *ZpkgManifestCommand {
	cmd := &ZpkgManifestCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "manifest [ZPKG PATH]"
	cmd.Short = "Dump ZPKG file manifest"
	cmd.Long = "Dump ZPKG file manifest"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgManifestCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgManifestCommand) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	manager := zpkg.NewManager()

	SetupEventHandlers(manager.Emitter, z.Ui)

	output, err := manager.Manifest(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(output)

	return err
}
