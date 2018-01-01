package commands

import (
	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/config"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmRemoveCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRemoveCommand() *ZpmRemoveCommand {
	cmd := &ZpmRemoveCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "remove"
	cmd.Short = "Remove packages from a ZPS image"
	cmd.Long = "remove packages from a ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmRemoveCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRemoveCommand) run(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to remove")
	}

	// Load config
	cfg, err := config.LoadConfig(root, image)
	if err != nil {
		z.Fatal(err.Error())
	}

	// Setup ZPM transaction
	transaction := zpm.NewTransaction(cfg.CurrentImage.Path, &zpm.Db{cfg.DbPath()},"remove")
	transaction.On("info", func(msg string) {
		z.Info(msg)
	})
	transaction.On("warn", func(msg string) {
		z.Warn(msg)
	})

	for _, arg := range args {
		transaction.AddPackage(arg)
	}

	err = transaction.Realize()
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
