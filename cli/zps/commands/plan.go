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
	"errors"

	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpsPlanCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPlanCommand() *ZpsPlanCommand {
	cmd := &ZpsPlanCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "plan"
	cmd.Short = "Plan ZPS transaction"
	cmd.Long = "Plan ZPS transaction"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsPlanCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPlanCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	if cmd.Flags().Arg(0) == "" {
		return errors.New("plan action required")
	}

	if cmd.Flags().Arg(1) == "" {
		return errors.New("at least one package must be specified")
	}

	_, err = mgr.Plan(cmd.Flags().Arg(0), cmd.Flags().Args()[1:])
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
