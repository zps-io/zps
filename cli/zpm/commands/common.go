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

	"github.com/fezz-io/zps/cli"
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

	emitter.On("manager.error", func(message string) {
		ui.Error(fmt.Sprint("x ", message))
	})

	emitter.On("manager.info", func(message string) {
		ui.Info(fmt.Sprint("* ", message))
	})

	emitter.On("manager.out", func(message string) {
		ui.Info(message)
	})

	emitter.On("manager.warn", func(message string) {
		ui.Warn(fmt.Sprint("~ ", message))
	})

	emitter.On("manager.fetch", func(message string) {
		ui.Info(fmt.Sprint("* ", message))
	})

	emitter.On("manager.freeze", func(message string) {
		ui.Blue(fmt.Sprint("* ", message))
	})

	emitter.On("manager.refresh", func(message string) {
		ui.Info(fmt.Sprint("* ", message))
	})

	emitter.On("manager.thaw", func(message string) {
		ui.Yellow(fmt.Sprint("* ", message))
	})

	emitter.On("publisher.publish", func(message string) {
		ui.Info(fmt.Sprint("* published ", message))
	})

	emitter.On("publisher.channel", func(message string) {
		ui.Info(fmt.Sprint("* channel", message))
	})

	emitter.On("transaction.noop", func(message string) {
		ui.Warn(fmt.Sprint("> ", message))
	})

	emitter.On("transaction.install", func(message string) {
		ui.Info(fmt.Sprint("+ ", message))
	})

	emitter.On("transaction.remove", func(message string) {
		ui.Red(fmt.Sprint("- ", message))
	})
}
