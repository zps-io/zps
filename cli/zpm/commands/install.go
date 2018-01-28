package commands

import (

	"errors"

	"github.com/solvent-io/zps/cli"

	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
	"fmt"
)

type ZpmInstallCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmInstallCommand() *ZpmInstallCommand {
	cmd := &ZpmInstallCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "install"
	cmd.Short = "Install packages into ZPS image"
	cmd.Long = "Install packages into ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmInstallCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmInstallCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to install")
	}

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("fetch", func(fetch string) {
		z.Info(fmt.Sprint("* fetching -> ", fetch))
	})

	mgr.On("install", func(install string) {
		z.Info(install)
	})

	mgr.On("remove", func(remove string) {
		z.Info(remove)
	})

	err = mgr.Install(args)
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
