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
	"fmt"

	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpsRepoInitCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRepoInitCommand() *ZpsRepoInitCommand {
	cmd := &ZpsRepoInitCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "init [REPO_NAME]"
	cmd.Short = "Initialize a ZPS repository"
	cmd.Long = "Initialize a ZPS repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().Bool("yes", false, "Do not ask for confirmation")

	return cmd
}

func (z *ZpsRepoInitCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRepoInitCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	repoName := cmd.Flags().Arg(0)

	if repoName == "" {
		return errors.New("Repo name required")
	}

	isSilent, _ := cmd.Flags().GetBool("yes")

	if !isSilent {
		fmt.Printf("ZPS will delete all data in repository '%s', proceed? [y/n]: ", repoName)

		var response string
		fmt.Scanln(&response)

		switch response {
		case "y", "Y", "Yes", "yes":
			fmt.Println("Initializing repo...")
		case "n", "N", "No", "no":
			return nil
		}
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.RepoInit(repoName)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
