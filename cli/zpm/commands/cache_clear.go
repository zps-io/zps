package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmCacheClearCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmCacheClearCommand() *ZpmCacheClearCommand {
	cmd := &ZpmCacheClearCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "clear"
	cmd.Short = "Clear ZPS image cache"
	cmd.Long = "Clear ZPS image cache"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmCacheClearCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmCacheClearCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("clear", func(clear string) {
		z.Info(clear)
	})

	err = mgr.Clear()
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
