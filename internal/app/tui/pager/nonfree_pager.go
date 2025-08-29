package pager

// // SPDX-License-Identifier: GPL-3.0-or-later
// //
// // Stapler
// // Copyright (C) 2025 The Stapler Authors
// //
// // This program is free software: you can redistribute it and/or modify
// // it under the terms of the GNU General Public License as published by
// // the Free Software Foundation, either version 3 of the License, or
// // (at your option) any later version.
// //
// // This program is distributed in the hope that it will be useful,
// // but WITHOUT ANY WARRANTY; without even the implied warranty of
// // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// // GNU General Public License for more details.
// //
// // You should have received a copy of the GNU General Public License
// // along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leonelquinteros/gotext"
)

type NonfreePager struct {
	model tea.Model
}

var urlBoxStyle = func() lipgloss.Style {
	return DefaultTitleStyle.BorderStyle(lipgloss.RoundedBorder())
}()

func NewNonfree(name, content, url string) *NonfreePager {
	cfg := PagerConfig{
		Name:     name,
		Content:  content,
		URL:      url,
		Title:    DefaultTitleStyle,
		Info:     DefaultInfoStyle,
		Help:     DefaultHelpStyle,
		URLStyle: DefaultURLStyle,
		Controls: []string{
			gotext.Get("y - Accept"),
			gotext.Get("n/q/Esc - Decline"),
		},
		KeyHandler: func(msg tea.KeyMsg, m *GenericModel) (bool, tea.Cmd) {
			switch msg.String() {
			case "y":
				v := true
				m.accepted = &v
				return true, tea.Quit
			case "n", "q", "esc", "ctrl+c":
				v := false
				m.accepted = &v
				return true, tea.Quit
			}
			return false, nil
		},
		HeaderRenderer: func(m GenericModel) string {
			warning := DefaultWarningStyle.Render("[!] " + gotext.Get("NON-FREE SOFTWARE AGREEMENT") + " [!]")
			title := DefaultTitleStyle.Render(fmt.Sprintf("%s - %s", warning, m.cfg.Name))
			line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
			return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
		},
		FooterRenderer: func(m GenericModel) string {
			info := m.cfg.Info.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
			line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))

			helpLine := ""
			if len(m.cfg.Controls) > 0 {
				helpLine = m.cfg.Help.Render(strings.Join(m.cfg.Controls, " • "))
			}

			urlLine := ""
			if m.cfg.URL != "" {
				urlLine = urlBoxStyle.Render(gotext.Get("Full license available at: %s", m.cfg.URLStyle.Render(m.cfg.URL)))
			}

			return fmt.Sprintf("%s\n%s\n%s", urlLine, lipgloss.JoinHorizontal(lipgloss.Center, line, info), helpLine)
		},
	}

	return &NonfreePager{
		model: New(cfg),
	}
}

func (p *NonfreePager) Run() (bool, error) {
	prog := tea.NewProgram(
		p.model,
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	)

	finalModel, err := prog.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(GenericModel); ok {
		return *m.accepted, nil
	}

	return false, nil
}
