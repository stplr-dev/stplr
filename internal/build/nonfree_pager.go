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

package build

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leonelquinteros/gotext"
)

var (
	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5733")).
			Bold(true)

	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
			Underline(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()

	urlBoxStyle = func() lipgloss.Style {
		return titleStyle.BorderStyle(lipgloss.RoundedBorder())
	}()
)

type NonfreePager struct {
	model nonfreeModel
}

func NewNonfree(name, content, url string) *NonfreePager {
	return &NonfreePager{
		model: nonfreeModel{
			name:     name,
			content:  content,
			url:      url,
			accepted: false,
		},
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

	if m, ok := finalModel.(nonfreeModel); ok {
		return m.accepted, nil
	}

	return false, nil
}

type nonfreeModel struct {
	name     string
	content  string
	url      string
	ready    bool
	viewport viewport.Model
	accepted bool
}

func (nm nonfreeModel) Init() tea.Cmd {
	return nil
}

func (pm nonfreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		switch k {
		case "ctrl+c", "q", "esc", "n":
			pm.accepted = false
			return pm, tea.Quit
		case "y":
			pm.accepted = true
			return pm, tea.Quit
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(pm.headerView())
		footerHeight := lipgloss.Height(pm.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !pm.ready {
			pm.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			pm.viewport.HighPerformanceRendering = true
			pm.viewport.YPosition = headerHeight + 1

			fullContent := "\n" + pm.content

			pm.viewport.SetContent(fullContent)
			pm.ready = true
		} else {
			pm.viewport.Width = msg.Width
			pm.viewport.Height = msg.Height - verticalMarginHeight
		}

		cmds = append(cmds, viewport.Sync(pm.viewport))
	}

	// Handle keyboard and mouse events in the viewport
	pm.viewport, cmd = pm.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return pm, tea.Batch(cmds...)
}

func (m nonfreeModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m nonfreeModel) headerView() string {
	warning := warningStyle.Render("[!] " + gotext.Get("NON-FREE SOFTWARE AGREEMENT") + " [!]")
	title := titleStyle.Render(fmt.Sprintf("%s - %s", warning, m.name))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (nm nonfreeModel) footerView() string {
	controls := []string{
		gotext.Get("y - Accept"),
		gotext.Get("n/q/Esc - Decline"),
	}

	info := infoStyle.Render(fmt.Sprintf("%3.f%%", nm.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, nm.viewport.Width-lipgloss.Width(info)))
	bottomLine := helpStyle.Render(strings.Join(controls, " • "))

	optionalUrl := ""

	if nm.url != "" {
		optionalUrl = urlBoxStyle.Render(gotext.Get("Full license available at: %s", urlStyle.Render(nm.url)))
	}

	return fmt.Sprintf(
		"%s\n%s\n%s",
		optionalUrl,
		lipgloss.JoinHorizontal(lipgloss.Center, line, info),
		bottomLine,
	)
}
