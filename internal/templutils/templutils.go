// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package templutils

import (
	"os"
	"strings"
	"text/template"

	"go.alt-gnome.ru/x/appstream"
)

func localizedText(m appstream.LocalizedMap, langs ...string) string {
	if len(langs) == 0 {
		langs = []string{os.Getenv("LANG"), "С", "en", ""}
	}

	for _, lang := range langs {
		lang = strings.SplitN(strings.SplitN(lang, ".", 2)[0], "_", 2)[0]
		for _, t := range m {
			if t.Lang == lang {
				return t.Value
			}
		}
	}
	return ""
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = "  " + line
		}
	}
	return strings.Join(lines, "\n")
}

var commonFuncs = template.FuncMap{
	"localized": localizedText,
	"indent":    indent,
}

func NewPackageTemplate() *template.Template {
	return template.New("").Funcs(commonFuncs)
}
