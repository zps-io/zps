package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpmTransactionCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmTransactionCommand() *ZpmTransactionCommand {
	cmd := &ZpmTransactionCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "transaction"
	cmd.Short = "Work with ZPS image transactions"
	cmd.Long = "Work with ZPS image transactions"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpmTransactionListCommand().Command)
	return cmd
}

func (z *ZpmTransactionCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmTransactionCommand) run(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
