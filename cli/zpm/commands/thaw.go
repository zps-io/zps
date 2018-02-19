package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"

	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmThawCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmThawCommand() *ZpmThawCommand {
	cmd := &ZpmThawCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "thaw"
	cmd.Short = "Un-freeze package version in ZPS image"
	cmd.Long = "Un-freeze package version in ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmThawCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmThawCommand) run(cmd *cobra.Command, args []string) error {
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

	mgr.On("thaw", func(thaw string) {
		z.Yellow(thaw)
	})

	err = mgr.Thaw(args)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
