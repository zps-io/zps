package commands

import (
	"net/url"

	"path/filepath"

	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/config"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
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
	root, _ := cmd.Flags().GetString("root")
	image, _ := cmd.Flags().GetString("image")

	if cmd.Flags().NArg() == 0 {
		return errors.New("Must provide at least one package uri to install")
	}

	// Load config
	cfg, err := config.LoadConfig(root, image)
	if err != nil {
		z.Fatal(err.Error())
	}

	// Setup ZPM transaction
	transaction := zpm.NewTransaction(cfg.CurrentImage.Path, &zpm.Db{cfg.DbPath()}, "install")
	transaction.On("info", func(msg string) {
		z.Info(msg)
	})
	transaction.On("warn", func(msg string) {
		z.Warn(msg)
	})

	for _, arg := range args {
		file, err := filepath.Abs(arg)

		uri, err := url.Parse(file)
		if err != nil {
			z.Fatal(err.Error())
		}

		transaction.AddUri(uri)
	}

	// Run ZPM transaction
	err = transaction.Realize()
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
