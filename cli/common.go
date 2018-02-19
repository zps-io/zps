package cli

import (
	"log"
	"os"

	"github.com/mitchellh/colorstring"
)

type Ui struct {
	color bool

	out *log.Logger

	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewUi() *Ui {
	ui := &Ui{}
	ui.color = true

	ui.out = log.New(os.Stdout, "", 0)

	ui.debug = log.New(os.Stdout, "", 0)
	ui.info = log.New(os.Stdout, "", 0)
	ui.warn = log.New(os.Stdout, "", 0)
	ui.error = log.New(os.Stderr, "", 0)

	return ui
}

func (u *Ui) NoColor(disable bool) *Ui {
	if disable {
		u.color = false
	} else {
		u.color = true
	}

	return u
}

func (u *Ui) Colorize(list []string) []string {
	for index := range list {
		list[index] = colorstring.Color(list[index])
	}

	return list
}

func (u *Ui) Out(content string) {
	u.out.Print(content)
}

func (u *Ui) Debug(content string) {
	if u.color {
		u.info.Println(colorstring.Color("[magenta]" + content))
	} else {
		u.info.Println(content)
	}
}

func (u *Ui) Info(content string) {
	if u.color {
		u.info.Println(colorstring.Color("[green]" + content))
	} else {
		u.info.Println(content)
	}
}

func (u *Ui) Warn(content string) {
	if u.color {
		u.warn.Println(colorstring.Color("[yellow]" + content))
	} else {
		u.warn.Println(content)
	}
}

func (u *Ui) Error(content string) {
	if u.color {
		u.error.Println(colorstring.Color("[red]" + content))
	} else {
		u.error.Println(content)
	}
}

func (u *Ui) Fatal(content string) {
	u.Error(content)
	os.Exit(1)
}
