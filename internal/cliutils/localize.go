// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
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
	"github.com/AlecAivazis/survey/v2"
	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"
)

// Make the application more internationalized
func Localize(app *cli.Command) {
	cli.RootCommandHelpTemplate = GetAppCliTemplate()
	cli.CommandHelpTemplate = GetCommandHelpTemplate()
	cli.SubcommandHelpTemplate = GetSubcommandHelpTemplate()
	cli.UsageCommandHelp = gotext.Get("Shows a list of commands or help for one command")
	cli.HelpFlag.(*cli.BoolFlag).Usage = gotext.Get("Show help")

	survey.MultiSelectQuestionTemplate = GetMultiSelectQuestionTemplate()
}
