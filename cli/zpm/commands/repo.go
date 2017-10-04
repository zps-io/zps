package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpmRepoCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRepoCommand() *ZpmRepoCommand {
	cmd := &ZpmRepoCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "repo"
	cmd.Short = "Work with ZPS repositories"
	cmd.Long = "Work with ZPS repositories"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.AddCommand(NewZpmRepoInitCommand().Command)
	cmd.AddCommand(NewZpmRepoListCommand().Command)
	return cmd
}

func (z *ZpmRepoCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRepoCommand) run(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
