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
	"github.com/fezz-io/zps/cli"

	"github.com/fezz-io/zps/zpm"

	"errors"

	"github.com/spf13/cobra"
)

type ZpsTplCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsTplCommand() *ZpsTplCommand {
	cmd := &ZpsTplCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "tpl [TEMPLATE_FILE]"
	cmd.Short = "Process a template and write to standard out"
	cmd.Long = "Process a template and write to standard out"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("profile", "default", "Profile for configuration context")

	return cmd
}

func (z *ZpsTplCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsTplCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	profile, _ := cmd.Flags().GetString("profile")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide the path to a template file")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Tpl(cmd.Flags().Arg(0), profile)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
