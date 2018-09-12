package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpkgRootCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgRootCommand() *ZpkgRootCommand {
	cmd := &ZpkgRootCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zpkg"
	cmd.Short = "ZPKG work with ZPKG files"
	cmd.Long = "ZPKG work with ZPKG files"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.PersistentFlags().Bool("no-color", false, "Disable color")

	cmd.AddCommand(NewZpkgBuildCommand().Command)

	return cmd
}

func (z *ZpkgRootCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgRootCommand) run(cmd *cobra.Command, args []string) error {
	z.Help()
	return nil
}
