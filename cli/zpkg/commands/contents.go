package commands

import (
	"sort"

	"errors"

	"github.com/ryanuber/columnize"
	"github.com/solvent-io/zps/action"
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpkgContentsCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgContentsCommand() *ZpkgContentsCommand {
	cmd := &ZpkgContentsCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "contents [ZPKG path]"
	cmd.Short = "List contents of a ZPKG"
	cmd.Long = "List contents of a ZPKG"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgContentsCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgContentsCommand) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NArg() != 1 {
		return errors.New("ZPKG Filename required")
	}

	reader := zpkg.NewReader(cmd.Flags().Arg(0), "")

	err := reader.Read()
	if err != nil {
		z.Fatal(err.Error())
	}

	var contents action.Actions
	contents = reader.Manifest.Section("Dir", "SymLink", "File")

	sort.Sort(contents)

	var output []string
	for _, fsObject := range contents {
		output = append(output, fsObject.Columns())
	}

	z.Out(columnize.SimpleFormat(output))

	return err
}
