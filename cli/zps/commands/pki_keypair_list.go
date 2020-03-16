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
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type ZpsPkiKeyPairListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiKeyPairListCommand() *ZpsPkiKeyPairListCommand {
	cmd := &ZpsPkiKeyPairListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List keypairs within ZPS pki store"
	cmd.Long = "List keypairs within ZPS pki store"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsPkiKeyPairListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiKeyPairListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	list, err := mgr.PkiKeyPairList()
	if err != nil {
		z.Fatal(err.Error())
	}

	if list != nil {
		z.Out(columnize.SimpleFormat(z.Colorize(list)))
	}

	return nil
}
