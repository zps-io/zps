/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2020 Zachary Schneider
 */

package commands

import (
	"errors"

	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpsPkiTrustFetchCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiTrustFetchCommand() *ZpsPkiTrustFetchCommand {
	cmd := &ZpsPkiTrustFetchCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "fetch [CERT_URI]"
	cmd.Short = "Fetch trusted certificates into the pki store"
	cmd.Long = "Fetch trusted certificates into the pki store"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("type", "user", "cerificate type: user|intermediate|ca")

	return cmd
}

func (z *ZpsPkiTrustFetchCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiTrustFetchCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("cert URI required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.PkiTrustFetch(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
