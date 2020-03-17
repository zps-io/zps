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

type ZpsPkiKeyPairImportCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsPkiKeyPairImportCommand() *ZpsPkiKeyPairImportCommand {
	cmd := &ZpsPkiKeyPairImportCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "import [CERT_FILE] [KEY_FILE]"
	cmd.Short = "Import signing key pair into ZPS pki store"
	cmd.Long = "Import signing key pair into ZPS pki store"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsPkiKeyPairImportCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsPkiKeyPairImportCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("cert file name required")
	}

	if cmd.Flags().Arg(1) == "" {
		return errors.New("key file name required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.PkiKeyPairImport(cmd.Flags().Arg(0), cmd.Flags().Arg(1))
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
