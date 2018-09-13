package commands

import (
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
	"github.com/spf13/cobra"
)

type ZpkgBuildCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgBuildCommand() *ZpkgBuildCommand {
	cmd := &ZpkgBuildCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "build [ZPKGFILE PATH]"
	cmd.Short = "Build a ZPKG from a Zpkgfile"
	cmd.Long = "Build a ZPKG from a Zpkgfile"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.Flags().String("target-path", "", "Target path for included file system objects")
	cmd.Flags().String("work-path", "", "Work path for ZPKG creation")
	cmd.Flags().String("output-path", "", "Output path for ZPKG")
	cmd.Flags().Bool("restrict", false, "Restrict included filesystem objects to those present in Zpkgfile")
	cmd.Flags().Bool("secure", false, "Ensure filesystem objects are super user owned")

	return cmd
}

func (z *ZpkgBuildCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgBuildCommand) run(cmd *cobra.Command, args []string) error {
	targetPath, _ := cmd.Flags().GetString("target-path")
	outputPath, _ := cmd.Flags().GetString("output-path")
	workPath, _ := cmd.Flags().GetString("work-path")
	restrict, _ := cmd.Flags().GetBool("restrict")
	secure, _ := cmd.Flags().GetBool("secure")

	builder := zpkg.NewBuilder()

	SetupEventHandlers(builder, z.Ui)

	builder.ZpfPath(cmd.Flags().Arg(0)).TargetPath(targetPath).WorkPath(workPath).OutputPath(outputPath).Restrict(restrict).Secure(secure)

	_, err := builder.Build()

	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
