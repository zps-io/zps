package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/spf13/cobra"
)

type ZpmRootCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmRootCommand() *ZpmRootCommand {
	cmd := &ZpmRootCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zpm"
	cmd.Short = "ZPM is the package management component of ZPS the Z Package System"
	cmd.Long = "ZPM is the package management component of ZPS the Z Package System"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.PersistentFlags().Bool("no-color", false, "Disable color")
	cmd.PersistentFlags().String("image", "", "ZPS image name/id")

	cmd.AddCommand(NewZpmCacheCommand().Command)
	cmd.AddCommand(NewZpmContentsCommand().Command)
	cmd.AddCommand(NewZpmFreezeCommand().Command)
	cmd.AddCommand(NewZpmInfoCommand().Command)
	cmd.AddCommand(NewZpmInstallCommand().Command)
	cmd.AddCommand(NewZpmListCommand().Command)
	cmd.AddCommand(NewZpmPlanCommand().Command)
	cmd.AddCommand(NewZpmPublishCommand().Command)
	cmd.AddCommand(NewZpmRefreshCommand().Command)
	cmd.AddCommand(NewZpmRemoveCommand().Command)
	cmd.AddCommand(NewZpmRepoCommand().Command)
	cmd.AddCommand(NewZpmStatusCommand().Command)
	cmd.AddCommand(NewZpmThawCommand().Command)
	cmd.AddCommand(NewZpmTransactionCommand().Command)
	cmd.AddCommand(NewZpmVersionCommand().Command)

	return cmd
}

func (z *ZpmRootCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmRootCommand) run(cmd *cobra.Command, args []string) error {
	z.Help()
	return nil
}
