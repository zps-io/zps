package cli

import (
	"log"
	"os"

	"github.com/mitchellh/colorstring"
)

type Ui struct {
	out *log.Logger

	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger

	colorize *colorstring.Colorize
}

func NewUi() *Ui {
	ui := &Ui{}

	ui.out = log.New(os.Stdout, "", 0)

	ui.debug = log.New(os.Stdout, "", 0)
	ui.info = log.New(os.Stdout, "", 0)
	ui.warn = log.New(os.Stdout, "", 0)
	ui.error = log.New(os.Stderr, "", 0)

	ui.colorize = &colorstring.Colorize{Colors: colorstring.DefaultColors}
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
	u.info.Println(u.colorize.Color("[magenta]" + content))
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
