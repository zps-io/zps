package commands

import (
	"fmt"

	"github.com/solvent-io/zps/cli"
	"github.com/solvent-io/zps/zpkg"
)

func SetupEventHandlers(builder *zpkg.Builder, ui *cli.Ui) {
	builder.On("action.info", func(message string) {
		ui.Info(fmt.Sprint("* ", message))
	})

	builder.On("action.debug", func(message string) {
		ui.Debug(fmt.Sprint("> ", message))
	})

	builder.On("action.error", func(message string) {
		ui.Error(fmt.Sprint("x ", message))
	})

	builder.On("action.warn", func(message string) {
		ui.Warn(fmt.Sprint("~ ", message))
	})

	builder.On("builder.complete", func(message string) {
		ui.Warn(fmt.Sprint("=> ", message))
	})
}
