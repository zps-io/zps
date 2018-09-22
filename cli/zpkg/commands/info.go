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

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpkgInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgInfoCommand() *ZpkgInfoCommand {
	cmd := &ZpkgInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info [ZPKG path]"
	cmd.Short = "Display ZPKG file information"
	cmd.Long = "Display ZPKG file information"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgInfoCommand) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	manager := zpkg.NewManager()

	SetupEventHandlers(manager.Emitter, z.Ui)

	output, err := manager.Info(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(output)

	return err
}
