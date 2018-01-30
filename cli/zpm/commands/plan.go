package commands

import (
	"errors"
	"fmt"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpm"
	"github.com/spf13/cobra"
)

type ZpmPlanCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpmPlanCommand() *ZpmPlanCommand {
	cmd := &ZpmPlanCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "plan"
	cmd.Short = "Plan ZPM transaction"
	cmd.Long = "Plan ZPM transaction"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpmPlanCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpmPlanCommand) run(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	// Load manager
	mgr, err := zpm.NewManager(image)
	if err != nil {
		z.Fatal(err.Error())
	}

	mgr.On("install", func(pkg string) {
		z.Info(fmt.Sprint("+ ", pkg))
	})

	mgr.On("remove", func(pkg string) {
		z.Info(fmt.Sprint("[red]- ", pkg))
	})

	mgr.On("noop", func(pkg string) {
		z.Warn(fmt.Sprint("-> ", pkg))
	})

	if cmd.Flags().Arg(0) == "" {
		return errors.New("plan action required")
	}

	if cmd.Flags().Arg(1) == "" {
		return errors.New("at least one package must be specified")
	}

	_, err = mgr.Plan(cmd.Flags().Arg(0), cmd.Flags().Args()[1:])
	if err != nil {
		z.Fatal(err.Error())
	}

	return nil
}
