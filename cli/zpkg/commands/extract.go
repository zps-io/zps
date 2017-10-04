package commands

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"errors"

	"github.com/solvent-io/zps/action"
	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/provider"
	"github.com/solvent-io/zps/zpkg"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type ZpkgExtractCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpkgExtractCommand() *ZpkgExtractCommand {
	cmd := &ZpkgExtractCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "extract"
	cmd.Short = "Extract contents of a ZPKG"
	cmd.Long = "Extract contents of a ZPKG"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	return cmd
}

func (z *ZpkgExtractCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpkgExtractCommand) run(cmd *cobra.Command, args []string) error {
	var err error

	if cmd.Flags().NArg() == 0 {
		return errors.New("ZPKG Filename required")
	}

	extractPath := cmd.Flags().Arg(1)

	if extractPath == "" {
		extractPath, err = os.Getwd()
		if err != nil {
			z.Fatal(err.Error())
		}
	}

	reader := zpkg.NewReader(cmd.Flags().Arg(0), "")

	err = reader.Read()
	if err != nil {
		z.Fatal(err.Error())
	}

	ctx := action.GetContext(action.NewOptions(), reader.Manifest)
	ctx = context.WithValue(ctx, "payload", reader.Payload)
	ctx.Value("options").(*action.Options).TargetPath = extractPath

	var contents action.Actions
	contents = reader.Manifest.Section("dir", "symlink", "file")

	sort.Sort(contents)

	for _, fsObject := range contents {
		z.Info(fmt.Sprintf("Extracted => %s %s", strings.ToUpper(fsObject.Type()), path.Join(args[1], fsObject.Key())))

		err = provider.Get(fsObject).Realize("install", ctx)
		if err != nil {
			z.Fatal(err.Error())
		}
	}

	return err
}
