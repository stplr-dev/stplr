// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package pager

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leonelquinteros/gotext"
)

type Pager struct {
	model tea.Model
}

func NewCode(name, content string) *Pager {
	cfg := PagerConfig{
		Name:    name,
		Content: content,
		Title:   DefaultTitleStyle,
		Info:    DefaultInfoStyle,
		Help:    DefaultHelpStyle,
		Controls: []string{
			gotext.Get("q/Esc - Quit"),
		},
		KeyHandler: func(msg tea.KeyMsg, m *GenericModel) (bool, tea.Cmd) {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return true, tea.Quit
			}
			return false, nil
		},
	}
	return &Pager{
		New(cfg),
	}
}

func (p *Pager) Run() error {
	prog := tea.NewProgram(
		p.model,
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	)

	_, err := prog.Run()
	return err
}
