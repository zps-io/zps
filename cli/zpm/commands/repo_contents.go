package commands

import (
	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
	"errors"
)

type ZpmRepoContentsCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRepoContentsCommand() *ZpmRepoContentsCommand {
	cmd := &ZpmRepoContentsCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "contents"
	cmd.Short = "Show contents of a ZPM repository"
	cmd.Long = "Show contents of a ZPM repository"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRepoContentsCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRepoContentsCommand) run(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().Arg(0) == "" {
		return errors.New("Repo name required")
	}

	// Load manager
	mgr, err := zpm.NewManager(root, image)
	if err != nil {
		z.Fatal(err.Error())
	}

	contents, err := mgr.RepoContents(cmd.Flags().Arg(0))
	if err != nil {
		z.Fatal(err.Error())
	}
	if contents == nil {
		z.Warn("Repository is empty")
		return nil
	}

	z.Out(columnize.SimpleFormat(contents))

	return nil
}
