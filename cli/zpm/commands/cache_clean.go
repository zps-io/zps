package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmCacheCleanCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmCacheCleanCommand() *ZpmCacheCleanCommand {
	cmd := &ZpmCacheCleanCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "clean"
	cmd.Short = "Clean ZPS image cache"
	cmd.Long = "Clean ZPS image cache"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmCacheCleanCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmCacheCleanCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("clean", func(clean string) {
		z.Info(clean)
	})

	err = mgr.Clean()
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
