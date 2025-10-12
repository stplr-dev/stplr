// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
// Copyright (C) 2025 The Stapler Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cliutils

import (
	"fmt"

	"github.com/leonelquinteros/gotext"
)

// Templates are based on https://github.com/urfave/cli/blob/3b17080d70a630feadadd23dd036cad121dd9a50/template.go

//nolint:unused
var (
	helpNameTemplate    = `{{$v := offset .HelpName 6}}{{wrap .HelpName 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}`
	descriptionTemplate = `{{wrap .Description 3}}`
	authorsTemplate     = `{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}`
	visibleCommandTemplate = `{{ $cv := offsetCommands .VisibleCommands 5}}{{range .VisibleCommands}}
   {{$s := join .Names ", "}}{{$s}}{{ $sp := subtract $cv (offset $s 3) }}{{ indent $sp ""}}{{wrap .Usage $cv}}{{end}}`
	visibleCommandCategoryTemplate = `{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{template "visibleCommandTemplate" .}}{{end}}{{end}}`
	visibleFlagCategoryTemplate = `{{range .VisibleFlagCategories}}
   {{if .Name}}{{.Name}}

   {{end}}{{$flglen := len .Flags}}{{range $i, $e := .Flags}}{{if eq (subtract $flglen $i) 1}}{{$e}}
{{else}}{{$e}}
   {{end}}{{end}}{{end}}`
	visibleFlagTemplate = `{{range $i, $e := .VisibleFlags}}
   {{wrap $e.String 6}}{{end}}`
	copyrightTemplate = `{{wrap .Copyright 3}}`
)

func GetAppCliTemplate() string {
	return fmt.Sprintf(`%s:
	{{template "helpNameTemplate" .}}

%s:
	{{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[%s]{{end}}{{if .Commands}} %s [%s]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[%s...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

%s:
	{{.Version}}{{end}}{{end}}{{if .Description}}

%s:
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

%s{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

%s:{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

%s:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

%s:{{template "visibleFlagTemplate" .}}{{end}}{{if .Copyright}}

%s:
   {{template "copyrightTemplate" .}}{{end}}
`, gotext.Get("NAME"), gotext.Get("USAGE"), gotext.Get("global options"), gotext.Get("command"), gotext.Get("command options"), gotext.Get("arguments"), gotext.Get("VERSION"), gotext.Get("DESCRIPTION"), gotext.Get("AUTHOR"), gotext.Get("COMMANDS"), gotext.Get("GLOBAL OPTIONS"), gotext.Get("GLOBAL OPTIONS"), gotext.Get("COPYRIGHT"))
}

func GetCommandHelpTemplate() string {
	return fmt.Sprintf(`%s:
   {{template "helpNameTemplate" .}}

%s:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleFlags}} [%s]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[%s...]{{end}}{{end}}{{if .Category}}

%s:
   {{.Category}}{{end}}{{if .Description}}

%s:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

%s:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

%s:{{template "visibleFlagTemplate" .}}{{end}}
`, gotext.Get("NAME"),
		gotext.Get("USAGE"),
		gotext.Get("command options"),
		gotext.Get("arguments"),
		gotext.Get("CATEGORY"),
		gotext.Get("DESCRIPTION"),
		gotext.Get("OPTIONS"),
		gotext.Get("OPTIONS"),
	)
}

func GetSubcommandHelpTemplate() string {
	return fmt.Sprintf(`%s:
   {{template "helpNameTemplate" .}}

%s:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}%s [%s]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[%s...]{{end}}{{end}}{{if .Description}}

%s:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

%s:{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

%s:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

%s:{{template "visibleFlagTemplate" .}}{{end}}
`,
		gotext.Get("NAME"),
		gotext.Get("USAGE"),
		gotext.Get("command"),
		gotext.Get("command options"),
		gotext.Get("arguments"),
		gotext.Get("DESCRIPTION"),
		gotext.Get("COMMANDS"),
		gotext.Get("OPTIONS"),
		gotext.Get("OPTIONS"),
	)
}

func GetMultiSelectQuestionTemplate() string {
	return fmt.Sprintf(`{{- define "option"}}
    {{- if eq .SelectedIndex .CurrentIndex }}{{color .Config.Icons.SelectFocus.Format }}{{ .Config.Icons.SelectFocus.Text }}{{color "reset"}}{{else}} {{end}}
    {{- if index .Checked .CurrentOpt.Index }}{{color .Config.Icons.MarkedOption.Format }} {{ .Config.Icons.MarkedOption.Text }} {{else}}{{color .Config.Icons.UnmarkedOption.Format }} {{ .Config.Icons.UnmarkedOption.Text }} {{end}}
    {{- color "reset"}}
    {{- " "}}{{- .CurrentOpt.Value}}{{ if ne ($.GetDescription .CurrentOpt) "" }} - {{color "cyan"}}{{ $.GetDescription .CurrentOpt }}{{color "reset"}}{{end}}
{{end}}
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} %s{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }}{{ .FilterMessage }}{{color "reset"}}
{{- if .ShowAnswer}}{{color "cyan"}} {{.Answer}}{{color "reset"}}{{"\n"}}
{{- else }}
	{{- "  "}}{{- color "cyan"}}[%s, %s,{{- if not .Config.RemoveSelectAll }} %s,{{end}}{{- if not .Config.RemoveSelectNone }} %s,{{end}} %s{{- if and .Help (not .ShowHelp)}}, {{ .Config.HelpInput }} %s{{end}}]{{color "reset"}}
  {{- "\n"}}
  {{- range $ix, $option := .PageEntries}}
    {{- template "option" $.IterateOption $ix $option}}
  {{- end}}
{{- end}}`,
		gotext.Get("Help"),
		gotext.Get("Use arrows to move"),
		gotext.Get("space to select"),
		gotext.Get("<right> to all"),
		gotext.Get("<left> to none"),
		gotext.Get("type to filter"),
		gotext.Get("for more help"),
	)
}
