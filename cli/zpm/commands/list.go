package commands

import (
	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"

	"github.com/solvent-io/zps/zpm"

	"github.com/spf13/cobra"
)

type ZpmListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmListCommand() *ZpmListCommand {
	cmd := &ZpmListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List packages installed in ZPS image"
	cmd.Long = "List packages installed in ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	var err error

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	list, err := mgr.List()
	if err != nil {
		z.Fatal(err.Error())
	}

	if list != nil {
		z.Out(columnize.SimpleFormat(z.Colorize(list)))
	}

	return nil
}
