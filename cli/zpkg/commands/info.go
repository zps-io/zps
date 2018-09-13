package commands

import (
	"fmt"

	"errors"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
	"github.com/solvent-io/zps/zps"
	"github.com/spf13/cobra"
)

type ZpkgInfoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgInfoCommand() *ZpkgInfoCommand {
	cmd := &ZpkgInfoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "info [ZPKG path]"
	cmd.Short = "Display ZPKG file information"
	cmd.Long = "Display ZPKG file information"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgInfoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgInfoCommand) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	reader := zpkg.NewReader(cmd.Flags().Arg(0), "")

	err := reader.Read()
	if err != nil {
		z.Fatal(err.Error())
	}

	pkg, err := zps.NewPkgFromManifest(reader.Manifest)
	if err != nil {
		z.Fatal(err.Error())
	}

	z.Out(fmt.Sprint("Name: ", pkg.Name()))
	z.Out(fmt.Sprint("Publisher: ", pkg.Publisher()))
	z.Out(fmt.Sprint("Semver: ", pkg.Version().Semver.String()))
	z.Out(fmt.Sprint("Timestamp: ", pkg.Version().Timestamp))
	z.Out(fmt.Sprint("OS: ", pkg.Os()))
	z.Out(fmt.Sprint("Arch: ", pkg.Arch()))
	z.Out(fmt.Sprint("Summary: ", pkg.Summary()))
	z.Out(fmt.Sprint("Description: ", pkg.Description()))

	return err
}
