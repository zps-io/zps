package commands

import (
	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"

	"github.com/solvent-io/zps/zpm"

	"errors"

	"github.com/spf13/cobra"
)

type ZpmInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmInfoCommand() *ZpmInfoCommand {
	cmd := &ZpmInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info"
	cmd.Short = "Show installed package metadata"
	cmd.Long = "Show installed package metadata"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmInfoCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	var err error

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide a package uri to query")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	info, err := mgr.Info(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	if info != nil {
		z.Out(columnize.SimpleFormat(z.Colorize(info)))
	}

	return nil
}
