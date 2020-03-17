package commands

import (
	"errors"

	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpsChannelCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsChannelCommand() *ZpsChannelCommand {
	cmd := &ZpsChannelCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "channel"
	cmd.Short = "Add a package to a channel with a ZPS repository"
	cmd.Long = "Add a package to a channel with a ZPS repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsChannelCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsChannelCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	if cmd.Flags().Arg(1) == "" {
		return errors.New("Must specify a zpkg to add to a channel")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.Channel(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags().Arg(2))
	if err != nil {
		z.Fatal(err.Error())
	}
	return nil
}
