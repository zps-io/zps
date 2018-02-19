package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"

	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmFreezeCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmFreezeCommand() *ZpmFreezeCommand {
	cmd := &ZpmFreezeCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "freeze"
	cmd.Short = "Freeze package version in ZPS image"
	cmd.Long = "Freeze package version in ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmFreezeCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmFreezeCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to freeze")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("error", func(error string) {
		z.Error(error)
	})

	mgr.On("freeze", func(freeze string) {
		z.Blue(freeze)
	})

	err = mgr.Freeze(args)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
