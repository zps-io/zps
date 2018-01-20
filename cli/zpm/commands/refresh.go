package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
	"fmt"
)

type ZpmRefreshCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRefreshCommand() *ZpmRefreshCommand {
	cmd := &ZpmRefreshCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "refresh"
	cmd.Short = "Refresh ZPM repository metadata"
	cmd.Long = "Refresh ZPM repository metadata"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRefreshCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRefreshCommand) run(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(root, image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("refresh", func(uri string) {
		z.Info(fmt.Sprint("* refreshed ", uri))
	})

	err = mgr.Refresh()
	if err != nil {
		z.Fatal(err.Error())
	}
	return nil
}
