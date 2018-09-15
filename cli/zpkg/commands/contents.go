package commands

import (
	"errors"

	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpkgContentsCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgContentsCommand() *ZpkgContentsCommand {
	cmd := &ZpkgContentsCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "contents [ZPKG path]"
	cmd.Short = "List contents of a ZPKG"
	cmd.Long = "List contents of a ZPKG"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgContentsCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgContentsCommand) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	manager := zpkg.NewManager()

	SetupEventHandlers(manager.Emitter, z.Ui)

	output, err := manager.Contents(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(columnize.SimpleFormat(output))

	return err
}
