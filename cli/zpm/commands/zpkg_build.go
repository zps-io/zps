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
	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpmZpkgBuildCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmZpkgBuildCommand() *ZpmZpkgBuildCommand {
	cmd := &ZpmZpkgBuildCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "build [ZPKGFILE PATH]"
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

func (z *ZpmZpkgBuildCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmZpkgBuildCommand) run(cmd *cobra.Command, args []string) error {
	targetPath, _ := cmd.Flags().GetString("target-path")
	outputPath, _ := cmd.Flags().GetString("output-path")
	workPath, _ := cmd.Flags().GetString("work-path")
	restrict, _ := cmd.Flags().GetBool("restrict")
	secure, _ := cmd.Flags().GetBool("secure")

	builder := zpkg.NewBuilder()

	SetupEventHandlers(builder.Emitter, z.Ui)

	builder.ZpfPath(cmd.Flags().Arg(0)).
		TargetPath(targetPath).WorkPath(workPath).
		OutputPath(outputPath).Restrict(restrict).
		Secure(secure)

	_, err := builder.Build()

	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
