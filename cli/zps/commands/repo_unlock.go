package commands

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/fezz-io/zps/cli"
	"github.com/fezz-io/zps/zpm"
)

type ZpsRepoUnlockCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRepoUnlockCommand() *ZpsRepoUnlockCommand {
	cmd := &ZpsRepoUnlockCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()

	cmd.Use = "unlock [REPO_NAME]"
	cmd.Short = "Unlock a ZPS repository"
	cmd.Long = "Unlock a ZPS repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpsRepoUnlockCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRepoUnlockCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	SetupEventHandlers(mgr.Emitter, z.Ui)

	err = mgr.RepoUnlock(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil

}
