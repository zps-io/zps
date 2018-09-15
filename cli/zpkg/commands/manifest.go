package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpkgManifestCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgManifestCommand() *ZpkgManifestCommand {
	cmd := &ZpkgManifestCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "manifest [ZPKG PATH]"
	cmd.Short = "Dump ZPKG file manifest"
	cmd.Long = "Dump ZPKG file manifest"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgManifestCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgManifestCommand) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	manager := zpkg.NewManager()

	SetupEventHandlers(manager.Emitter, z.Ui)

	output, err := manager.Manifest(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(output)

	return err
}
