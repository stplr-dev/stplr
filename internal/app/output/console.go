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

package output

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/leonelquinteros/gotext"
)

type ConsoleOutput struct {
	mu     sync.Mutex
	styles ConsoleStyles
}

type ConsoleStyles struct {
	Info     lipgloss.Style
	InfoPre  lipgloss.Style
	WarnPre  lipgloss.Style
	Warn     lipgloss.Style
	ErrorPre lipgloss.Style
}

func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{
		styles: ConsoleStyles{
			InfoPre: lipgloss.NewStyle().
				SetString("--> ").
				Foreground(lipgloss.Color("35")),
			WarnPre: lipgloss.NewStyle().
				SetString(strings.ToUpper("warn")).
				Bold(true).
				MaxWidth(4).
				Foreground(lipgloss.Color("192")),
			Warn: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214")),
			ErrorPre: lipgloss.NewStyle().
				SetString(gotext.Get("ERROR")).
				Padding(0, 1, 0, 1).
				Background(lipgloss.Color("204")).
				Foreground(lipgloss.Color("0")),
		},
	}
}

func (out *ConsoleOutput) print(msg string) {
	out.mu.Lock()
	defer out.mu.Unlock()
	fmt.Fprintln(os.Stderr, msg)
}

func (out *ConsoleOutput) Info(msg string, args ...any) {
	out.print(out.styles.InfoPre.Render() + out.styles.Info.Render(fmt.Sprintf(msg, args...)))
}

func (out *ConsoleOutput) Warn(msg string, args ...any) {
	out.print(fmt.Sprintf(out.styles.Warn.Render(msg), args...))
}

func (out *ConsoleOutput) Error(msg string, args ...any) {
	out.print(out.styles.ErrorPre.Render() + " " + fmt.Sprintf(msg, args...))
}
