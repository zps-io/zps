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
	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
	"github.com/zps-io/zps/zpm"
)

type ZpsZpkgBuildCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsZpkgBuildCommand() *ZpsZpkgBuildCommand {
	cmd := &ZpsZpkgBuildCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "build [ZPKGFILE_PATH]"
	cmd.Short = "Build a ZPKG from a Zpkgfile"
	cmd.Long = "Build a ZPKG from a Zpkgfile"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("target-path", "", "Target path for included file system objects")
	cmd.Flags().String("work-path", "", "Work path for ZPKG creation")
	cmd.Flags().String("output-path", "", "Output path for ZPKG")
	cmd.Flags().Bool("restrict", false, "Restrict included filesystem objects to those present in Zpkgfile")
	cmd.Flags().Bool("secure", false, "Ensure filesystem objects are super user owned")

	return cmd
}

func (z *ZpsZpkgBuildCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsZpkgBuildCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	targetPath, _ := cmd.Flags().GetString("target-path")
	outputPath, _ := cmd.Flags().GetString("output-path")
	workPath, _ := cmd.Flags().GetString("work-path")
	restrict, _ := cmd.Flags().GetBool("restrict")
	secure, _ := cmd.Flags().GetBool("secure")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.ZpkgBuild(cmd.Flags().Arg(0), targetPath, workPath, outputPath, restrict, secure)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
