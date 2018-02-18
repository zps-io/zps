package commands

import (
	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmTransactionListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmTransactionListCommand() *ZpmTransactionListCommand {
	cmd := &ZpmTransactionListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List ZPS image transactions"
	cmd.Long = "List ZPS image transactions"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmTransactionListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmTransactionListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("warn", func(err string) {
		z.Warn(err)
	})

	transactions, err := mgr.TransActionList()
	if err != nil {
		z.Fatal(err.Error())
	}

	if transactions != nil {
		z.Info(columnize.SimpleFormat(transactions))
	}

	return nil
}
