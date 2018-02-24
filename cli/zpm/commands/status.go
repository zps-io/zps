package commands

import (
	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"

	"github.com/solvent-io/zps/zpm"

	"errors"

	"github.com/spf13/cobra"
)

type ZpmStatusCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmStatusCommand() *ZpmStatusCommand {
	cmd := &ZpmStatusCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "status"
	cmd.Short = "Show status of specified zpkg uri"
	cmd.Long = "Show status of specified zpkg uri"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmStatusCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmStatusCommand) run(cmd *cobra.Command, args []string) error {
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

	status, versions, err := mgr.Status(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	var output []string
	if status != "" {
		output = append(output, "Status:|" + status)
		output = append(output, "Versions:|")
		z.Out(columnize.SimpleFormat(z.Colorize(output)))
	}

	if len(versions) > 0 {
		z.Out(columnize.SimpleFormat(z.Colorize(versions)))
	}

	return nil
}
