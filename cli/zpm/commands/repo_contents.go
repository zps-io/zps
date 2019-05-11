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
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type ZpmRepoContentsCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRepoContentsCommand() *ZpmRepoContentsCommand {
	cmd := &ZpmRepoContentsCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "contents"
	cmd.Short = "Show contents of a ZPM repository"
	cmd.Long = "Show contents of a ZPM repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRepoContentsCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRepoContentsCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	contents, err := mgr.RepoContents(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}
	if contents == nil {
		z.Warn("Repository is empty")
		return nil
	}

	z.Out(columnize.SimpleFormat(contents))

	return nil
}
