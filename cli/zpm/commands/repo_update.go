package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmRepoUpdateCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRepoUpdateCommand() *ZpmRepoUpdateCommand {
	cmd := &ZpmRepoUpdateCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "update"
	cmd.Short = "Update a ZPM repository configuration"
	cmd.Long = "Update a ZPM repository configuration"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRepoUpdateCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRepoUpdateCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	err = mgr.RepoUpdate(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}

