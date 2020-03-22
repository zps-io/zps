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

type ZpsPublishCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPublishCommand() *ZpsPublishCommand {
	cmd := &ZpsPublishCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "publish [REPO_NAME] [PKG ...]"
	cmd.Short = "Publish ZPKG(s) to a repository"
	cmd.Long = "Publish ZPKG(s) to a repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsPublishCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPublishCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	if cmd.Flags().Arg(1) == "" {
		return errors.New("Must specify at least one zpkg to publish")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Publish(cmd.Flags().Arg(0), cmd.Flags().Args()[1:]...)
	if err != nil {
		z.Fatal(err.Error())
	}
	return nil
}
