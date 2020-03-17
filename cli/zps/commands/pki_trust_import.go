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

type ZpsPkiTrustImportCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiTrustImportCommand() *ZpsPkiTrustImportCommand {
	cmd := &ZpsPkiTrustImportCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "import [CERT_FILE]"
	cmd.Short = "Import trusted certificate into ZPS pki store"
	cmd.Long = "Import trusted certificate into ZPS pki store"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("type", "user", "cerificate type: user|intermediate|ca")

	return cmd
}

func (z *ZpsPkiTrustImportCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiTrustImportCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	typ, _ := cmd.Flags().GetString("type")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("cert file name required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.PkiTrustImport(cmd.Flags().Arg(0), typ)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
