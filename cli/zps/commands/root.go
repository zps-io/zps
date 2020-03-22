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
	"github.com/fezz-io/zps/cli"
	"github.com/spf13/cobra"
)

// TODO get rid of this once cobra cmd has command groups
var UsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if and .HasAvailableSubCommands (eq .Name "zps")}}

Manage Current Image:{{range .Commands}}{{if and .IsAvailableCommand (IsManageCmd .Name)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Package Publishing/Fetching:{{range .Commands}}{{if and .IsAvailableCommand (IsPublishCmd .Name)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Images and ZPKGs:{{range .Commands}}{{if and .IsAvailableCommand (IsImgZpkgCmd .Name)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

ZPS:{{range .Commands}}{{if or (eq .Name "help") (and .IsAvailableCommand (IsZpsCmd .Name))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if and .HasAvailableSubCommands (ne .Name "zps")}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var ManageCommands = map[string]bool{
	"cache":       true,
	"contents":    true,
	"configure":   true,
	"freeze":      true,
	"info":        true,
	"install":     true,
	"list":        true,
	"pki":         true,
	"plan":        true,
	"refresh":     true,
	"remove":      true,
	"repo":        true,
	"status":      true,
	"thaw":        true,
	"transaction": true,
	"update":      true,
}

var PublishCommands = map[string]bool{
	"channel": true,
	"fetch":   true,
	"publish": true,
}

var ImgPkgCommands = map[string]bool{
	"image": true,
	"zpkg":  true,
}

var ZpsCommands = map[string]bool{
	"help":    true,
	"tpl":     true,
	"version": true,
}

type ZpsRootCommand struct {
	*cobra.Command
	*cli.Ui
}

func NewZpsRootCommand() *ZpsRootCommand {
	cmd := &ZpsRootCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zps"
	cmd.Short = "ZPS (Z Package System) The last word in package management"
	cmd.Long = "ZPS (Z Package System) The last word in package management"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	cmd.PersistentFlags().Bool("no-color", false, "Disable color")
	cmd.PersistentFlags().String("image", "", "ZPS image name/id")

	cmd.AddCommand(NewZpsCacheCommand().Command)
	cmd.AddCommand(NewZpsChannelCommand().Command)
	cmd.AddCommand(NewZpsContentsCommand().Command)
	cmd.AddCommand(NewZpsFetchCommand().Command)
	cmd.AddCommand(NewZpsFreezeCommand().Command)
	cmd.AddCommand(NewZpsImageCommand().Command)
	cmd.AddCommand(NewZpsInfoCommand().Command)
	cmd.AddCommand(NewZpsInstallCommand().Command)
	cmd.AddCommand(NewZpsListCommand().Command)
	cmd.AddCommand(NewZpsPkiCommand().Command)
	cmd.AddCommand(NewZpsPlanCommand().Command)
	cmd.AddCommand(NewZpsPublishCommand().Command)
	cmd.AddCommand(NewZpsRefreshCommand().Command)
	cmd.AddCommand(NewZpsRemoveCommand().Command)
	cmd.AddCommand(NewZpsRepoCommand().Command)
	cmd.AddCommand(NewZpsStatusCommand().Command)
	cmd.AddCommand(NewZpsThawCommand().Command)
	cmd.AddCommand(NewZpsTplCommand().Command)
	cmd.AddCommand(NewZpsTransactionCommand().Command)
	cmd.AddCommand(NewZpsUpdateCommand().Command)
	cmd.AddCommand(NewZpsVersionCommand().Command)
	cmd.AddCommand(NewZpsZpkgCommand().Command)

	cmd.SetUsageTemplate(UsageTemplate)
	cobra.AddTemplateFunc("IsManageCmd", cmd.isManageCmd)
	cobra.AddTemplateFunc("IsPublishCmd", cmd.isPublishCmd)
	cobra.AddTemplateFunc("IsImgZpkgCmd", cmd.isImgZpkgCmd)
	cobra.AddTemplateFunc("IsZpsCmd", cmd.isZpsCmd)

	return cmd
}

func (z *ZpsRootCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZpsRootCommand) run(cmd *cobra.Command, args []string) error {
	return z.Help()
}

func (z *ZpsRootCommand) isManageCmd(cmd string) bool {
	return ManageCommands[cmd]
}

func (z *ZpsRootCommand) isPublishCmd(cmd string) bool {
	return PublishCommands[cmd]
}

func (z *ZpsRootCommand) isImgZpkgCmd(cmd string) bool {
	return ImgPkgCommands[cmd]
}

func (z *ZpsRootCommand) isZpsCmd(cmd string) bool {
	return ZpsCommands[cmd]
}
