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
	"github.com/spf13/cobra"
)

type ZpsImageInitCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsImageInitCommand() *ZpsImageInitCommand {
	cmd := &ZpsImageInitCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "init [IMGFILE]"
	cmd.Short = "Initialize new ZPS image"
	cmd.Long = "Initialize new ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("arch", "", "select image arch [x86_64|arm64]")
	cmd.Flags().String("os", "", "select image os [darwin|linux]")
	cmd.Flags().String("name", "", "set image name")
	cmd.Flags().String("path", "", "override detected path")
	cmd.Flags().String("profile", "default", "select config profile")
	cmd.Flags().Bool("configure", false, "configure the image after init")
	cmd.Flags().Bool("force", false, "purge the image path, before init")
	cmd.Flags().Bool("helper", false, "reinstall ZPS helper")

	return cmd
}

func (z *ZpsImageInitCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsImageInitCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	arch, _ := cmd.Flags().GetString("arch")
	os, _ := cmd.Flags().GetString("os")
	name, _ := cmd.Flags().GetString("name")
	path, _ := cmd.Flags().GetString("path")
	profile, _ := cmd.Flags().GetString("profile")
	configure, _ := cmd.Flags().GetBool("configure")
	force, _ := cmd.Flags().GetBool("force")
	helper, _ := cmd.Flags().GetBool("helper")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.ImageInit(cmd.Flags().Arg(0), name, os, arch, path, profile, configure, force, helper)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
