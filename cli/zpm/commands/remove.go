package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmRemoveCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRemoveCommand() *ZpmRemoveCommand {
	cmd := &ZpmRemoveCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "remove"
	cmd.Short = "Remove packages from a ZPS image"
	cmd.Long = "remove packages from a ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRemoveCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRemoveCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to remove")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Remove(args)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
