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
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	"github.com/zps-io/zps/cli"
	"github.com/zps-io/zps/zpm"
)

type ZpsPkiTrustListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiTrustListCommand() *ZpsPkiTrustListCommand {
	cmd := &ZpsPkiTrustListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List trusted certificates within ZPS pki store"
	cmd.Long = "List trusted certificates within ZPS pki store"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsPkiTrustListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiTrustListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	list, err := mgr.PkiTrustList()
	if err != nil {
		z.Fatal(err.Error())
	}

	if list != nil {
		z.Out(columnize.SimpleFormat(z.Colorize(list)))
	}

	return nil
}
