/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package commands

import (
	"fmt"

	"github.com/chuckpreslar/emission"

	"github.com/solvent-io/zps/cli"
)

func SetupEventHandlers(emitter *emission.Emitter, ui *cli.Ui) {
	emitter.On("action.info", func(message string) {
		ui.Info(fmt.Sprint("* ", message))
	})

	emitter.On("action.debug", func(message string) {
		ui.Debug(fmt.Sprint("> ", message))
	})

	emitter.On("action.error", func(message string) {
		ui.Error(fmt.Sprint("x ", message))
	})

	emitter.On("action.warn", func(message string) {
		ui.Warn(fmt.Sprint("~ ", message))
	})

	emitter.On("builder.complete", func(message string) {
		ui.Warn(fmt.Sprint("=> ", message))
	})

	emitter.On("manager.info", func(message string) {
		ui.Info(message)
	})
}
