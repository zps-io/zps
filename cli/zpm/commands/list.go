package commands

import (
	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zps"
	"github.com/spf13/cobra"
	"github.com/solvent-io/zps/config"
	"github.com/solvent-io/zps/zpm"
)

type ZpmListCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmListCommand() *ZpmListCommand {
	cmd := &ZpmListCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "list"
	cmd.Short = "List packages installed in ZPS image"
	cmd.Long = "List packages installed in ZPS image"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmListCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmListCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	var err error

	// Load config
	cfg, err := config.LoadConfig(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	db := &zpm.Db{cfg.DbPath()}

	packages, err := db.Packages()
	if err != nil {
		z.Fatal(err.Error())
	}

	var output []string
	for _, manifest := range packages {
		pkg, _ := zps.NewPkgFromManifest(manifest)

		output = append(output, pkg.Columns())
	}

	if len(packages) == 0 {
		z.Warn("No packages installed.")
	} else {
		z.Out(columnize.SimpleFormat(output))
	}

	return nil
}
