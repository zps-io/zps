package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmPublishCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmPublishCommand() *ZpmPublishCommand {
	cmd := &ZpmPublishCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "publish"
	cmd.Short = "Publish to a ZPM repository"
	cmd.Long = "Publish to a ZPM repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmPublishCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmPublishCommand) run(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	if cmd.Flags().Arg(1) == "" {
		return errors.New("Must specify at least one zpkg to publish")
	}

	// Load manager
	mgr, err := zpm.NewManager(root, image)
	if err != nil {
		z.Fatal(err.Error())
	}

	err = mgr.Publish(cmd.Flags().Arg(0), cmd.Flags().Args()[1:]...)
	if err != nil {
		z.Fatal(err.Error())
	}
	return nil
}
