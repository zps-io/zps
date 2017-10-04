package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpmVersionCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmVersionCommand() *ZpmVersionCommand {
	cmd := &ZpmVersionCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "version"
	cmd.Short = "Print the version number of ZPM"
	cmd.Long = "Print the version number of ZPM"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmVersionCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmVersionCommand) run(cmd *cobra.Command, args []string) error {
	z.Out("ZPM v0.1.0")
	return nil
}
