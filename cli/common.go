/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package cli

import (
	"log"
	"os"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/mitchellh/colorstring"
)

type Ui struct {
	out *log.Logger

	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger

	spinner *wow.Wow

	colorize *colorstring.Colorize
}

func NewUi() *Ui {
	ui := &Ui{}

	ui.out = log.New(os.Stdout, "", 0)

	ui.debug = log.New(os.Stdout, "", 0)
	ui.info = log.New(os.Stdout, "", 0)
	ui.warn = log.New(os.Stdout, "", 0)
	ui.error = log.New(os.Stderr, "", 0)

	ui.colorize = &colorstring.Colorize{Colors: colorstring.DefaultColors, Reset: true}
	return ui
}

func (u *Ui) NoColor(disable bool) *Ui {

	u.colorize.Disable = disable

	return u
}

func (u *Ui) Colorize(list []string) []string {
	for index := range list {
		list[index] = u.colorize.Color(list[index])
	}

	return list
}

func (u *Ui) Out(content string) {
	u.out.Print(content)
}

func (u *Ui) Debug(content string) {
	u.info.Println(u.colorize.Color("[blue]" + content))
}

func (u *Ui) Info(content string) {
	u.info.Println(u.colorize.Color("[green]" + content))
}

func (u *Ui) Warn(content string) {
	u.warn.Println(u.colorize.Color("[yellow]" + content))
}

func (u *Ui) Error(content string) {
	u.error.Println(u.colorize.Color("[red]" + content))
}

func (u *Ui) Fatal(content string) {
	u.Error(content)
	os.Exit(1)
}

// Color shortcuts
func (u *Ui) Yellow(content string) {
	u.info.Println(u.colorize.Color("[yellow]" + content))
}

func (u *Ui) Blue(content string) {
	u.info.Println(u.colorize.Color("[blue]" + content))
}

func (u *Ui) Red(content string) {
	u.info.Println(u.colorize.Color("[red]" + content))
}

// Spin
func (u *Ui) Spin(content string) {
	u.spinner = wow.New(
		os.Stderr,
		spin.Spinner{
			Name:     1,
			Interval: 80,
			Frames: []string{
				u.colorize.Color(`[green]⠋`),
				u.colorize.Color(`[green]⠙`),
				u.colorize.Color(`[green]⠹`),
				u.colorize.Color(`[green]⠸`),
				u.colorize.Color(`[green]⠼`),
				u.colorize.Color(`[green]⠴`),
				u.colorize.Color(`[green]⠦`),
				u.colorize.Color(`[green]⠧`),
				u.colorize.Color(`[green]⠇`),
				u.colorize.Color(`[green]⠏`),
			},
		},
		u.colorize.Color("[green]"+content),
	)

	u.spinner.Start()
}

func (u *Ui) SpinSuccess(content string) {
	u.spinner.PersistWith(spin.Spinner{Frames: []string{""}}, u.colorize.Color("[green]"+content))
}

func (u *Ui) SpinError(content string) {
	u.spinner.PersistWith(spin.Spinner{Frames: []string{""}}, u.colorize.Color("[red]"+content))
}

func (u *Ui) SpinWarn(content string) {
	u.spinner.PersistWith(spin.Spinner{Frames: []string{""}}, u.colorize.Color("[yellow]"+content))
}
