package commands

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/chuckpreslar/emission"

	"errors"

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
	cmd.Use = "extract [ZPKG path] [extract path]"
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

	ctx := context.WithValue(context.Background(), "payload", reader.Payload)
	ctx = context.WithValue(ctx, "phase", "install")
	ctx = context.WithValue(ctx, "options", &provider.Options{TargetPath: extractPath})

	contents := reader.Manifest.Section("Dir", "SymLink", "File")

	sort.Sort(contents)

	factory := provider.DefaultFactory(emission.NewEmitter())

	for _, fsObject := range contents {
		z.Info(fmt.Sprintf("Extracted => %s %s", strings.ToUpper(fsObject.Type()), path.Join(args[1], fsObject.Key())))

		err = factory.Get(fsObject).Realize(ctx)
		if err != nil {
			z.Fatal(err.Error())
		}
	}

	return err
}
