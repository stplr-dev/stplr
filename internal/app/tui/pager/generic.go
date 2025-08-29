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
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	DefaultInfoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return DefaultTitleStyle.BorderStyle(b)
	}()

	DefaultHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666"))

	DefaultWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5733")).
				Bold(true)

	DefaultURLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
			Underline(true)
)

type PagerConfig struct {
	Name     string
	Content  string
	URL      string
	Title    lipgloss.Style
	Info     lipgloss.Style
	Help     lipgloss.Style
	Warning  lipgloss.Style
	URLStyle lipgloss.Style

	Controls []string

	KeyHandler     func(msg tea.KeyMsg, m *GenericModel) (bool, tea.Cmd)
	HeaderRenderer func(m GenericModel) string
	FooterRenderer func(m GenericModel) string
}

type GenericModel struct {
	cfg      PagerConfig
	viewport viewport.Model
	ready    bool
	accepted *bool // nil = no dialog, true/false = selection
}

func (m GenericModel) Init() tea.Cmd {
	return nil
}

func (m GenericModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.cfg.KeyHandler != nil {
			exit, extraCmd := m.cfg.KeyHandler(msg, &m)
			if extraCmd != nil {
				cmds = append(cmds, extraCmd)
			}
			if exit {
				return m, tea.Batch(cmds...)
			}
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.HighPerformanceRendering = true
			m.viewport.YPosition = headerHeight + 1
			m.viewport.SetContent("\n" + m.cfg.Content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
		cmds = append(cmds, viewport.Sync(m.viewport))
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m GenericModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m GenericModel) headerView() string {
	if m.cfg.HeaderRenderer != nil {
		return m.cfg.HeaderRenderer(m)
	} else {
		return DefaultHeader(m)
	}
}

func (m GenericModel) footerView() string {
	if m.cfg.FooterRenderer != nil {
		return m.cfg.FooterRenderer(m)
	} else {
		return DefaultFooter(m)
	}
}

func DefaultHeader(m GenericModel) string {
	title := m.cfg.Title.Render(m.cfg.Name)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func DefaultFooter(m GenericModel) string {
	info := m.cfg.Info.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))

	helpLine := ""
	if len(m.cfg.Controls) > 0 {
		helpLine = m.cfg.Help.Render(strings.Join(m.cfg.Controls, " • "))
	}

	return fmt.Sprintf("%s\n%s", lipgloss.JoinHorizontal(lipgloss.Center, line, info), helpLine)
}

func New(cfg PagerConfig) *GenericModel {
	return &GenericModel{cfg: cfg}
}
