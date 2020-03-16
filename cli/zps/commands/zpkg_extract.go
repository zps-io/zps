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
	"os"

	"github.com/fezz-io/zps/zpm"

	"github.com/fezz-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpsZpkgExtractCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsZpkgExtractCommand() *ZpsZpkgExtractCommand {
	cmd := &ZpsZpkgExtractCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "extract [ZPKG path] [extract path]"
	cmd.Short = "Extract contents of a ZPKG"
	cmd.Long = "Extract contents of a ZPKG"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsZpkgExtractCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsZpkgExtractCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	var err error

	if cmd.Flags().NArg() == 0 {
		return errors.New("ZPKG Filename required")
	}

	extractPath := cmd.Flags().Arg(1)

	if extractPath == "" {
		extractPath, err = os.Getwd()
		if err != nil {
			z.Fatal(err.Error())
		}
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.ZpkgExtract(cmd.Flags().Arg(0), extractPath)
	if err != nil {
		z.Fatal(err.Error())
	}

	return err
}
