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

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
	"github.com/zps-io/zps/zpm"
)

type ZpsRepoContentsCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRepoContentsCommand() *ZpsRepoContentsCommand {
	cmd := &ZpsRepoContentsCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "contents [REPO_NAME]"
	cmd.Short = "Show contents of a ZPS repository"
	cmd.Long = "Show contents of a ZPS repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsRepoContentsCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRepoContentsCommand) run(cmd *cobra.Command, args []string) error {
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
