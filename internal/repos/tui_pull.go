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

package repos

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.stplr.dev/stplr/pkg/types"
)

const (
	primaryColor     = lipgloss.Color("#00a8a3")
	errorColor       = lipgloss.Color("9")
	errorDarkerColor = lipgloss.Color("88")
)

var (
	textErorrStyle = lipgloss.NewStyle().
			Foreground(errorColor)
	textErorrDarkerStyle = lipgloss.NewStyle().
				Foreground(errorDarkerColor)
)

type progressViewportMsg struct {
	line     string
	isUpdate bool
}

type progressViewport struct {
	output        []string
	viewport      viewport.Model
	viewportReady bool
}

func (m progressViewport) Update(msg tea.Msg) (progressViewport, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.viewportReady {
			m.viewport.Width = msg.Width
			m.viewport.Height = 15
			m.viewportReady = true
		}
	case progressViewportMsg:
		if msg.isUpdate {
			if len(m.output) > 0 {
				m.output[len(m.output)-1] = msg.line
			} else {
				m.output = append(m.output, msg.line)
			}
		} else {
			m.output = append(m.output, msg.line)
		}

		m.viewport.SetContent(strings.Join(m.output, "\n"))
		m.viewport.GotoBottom()
		m.viewport.Height = len(m.output) + 3
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m progressViewport) Init() tea.Cmd {
	return nil
}

func (m progressViewport) View() string {
	return m.viewport.View()
}

type progressViewportWriter struct {
	onLine  func(line string, isUpdate bool)
	lineBuf []byte
}

func (w *progressViewportWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	for i := 0; i < len(p); i++ {
		b := p[i]

		switch b {
		case '\r':
			if len(w.lineBuf) > 0 {
				line := string(w.lineBuf)
				if w.onLine != nil {
					w.onLine(line, true)
				}
				w.lineBuf = w.lineBuf[:0]
			}

		case '\n':
			if len(w.lineBuf) > 0 {
				line := string(w.lineBuf)
				if w.onLine != nil {
					w.onLine(line, false)
				}
				w.lineBuf = w.lineBuf[:0]
			}

		default:
			w.lineBuf = append(w.lineBuf, b)
		}
	}

	return n, nil
}

//

// ================

type tryPullMsg struct{ err error }

type pullModel struct {
	ctx  context.Context
	done bool

	repo *types.Repo

	logs    []string
	urls    []string
	current int
	err     error
	status  string

	rs *Repos

	spinner spinner.Model
	gitBox  progressViewport

	writer             io.Writer
	updateRepoFromToml bool
}

func newPullModel(ctx context.Context, repo *types.Repo, rs *Repos, updateRepoFromToml bool) pullModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)
	urls := []string{repo.URL}
	urls = append(urls, repo.Mirrors...)

	gitBox := progressViewport{}
	gitBox.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	return pullModel{
		ctx:  ctx,
		repo: repo,
		rs:   rs,

		spinner: s,
		urls:    urls,
		status:  fmt.Sprintf("Pull %s", urls[0]),
		logs:    []string{},

		gitBox:             gitBox,
		updateRepoFromToml: updateRepoFromToml,
	}
}

func (m pullModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.tryPull(m.urls[m.current]))
}

func (m pullModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tryPullMsg:
		if msg.err != nil {
			line := textErorrDarkerStyle.
				Render(fmt.Sprintf("- Failed to pull from %s: %v", m.urls[m.current], strings.TrimSpace(msg.err.Error())))
			m.logs = append(m.logs, line)

			m.err = msg.err
			m.current++
			if m.current < len(m.urls) {
				m.status = fmt.Sprintf("Trying mirror %d: %s", m.current, m.urls[m.current])
				return m, tea.Batch(m.spinner.Tick, m.tryPull(m.urls[m.current]))
			}
			m.done = true
			m.status = textErorrStyle.Render("🞮 Failed to pull — the only source is unavailable")
			return m, tea.Quit
		}
		line := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render(fmt.Sprintf("- Pulled from %s", m.urls[m.current]))
		m.logs = append(m.logs, line)
		m.status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).Render("✔ Repository pulled successfully!")
		m.done = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.gitBox, cmd = m.gitBox.Update(msg)
	return m, cmd
}

func (m pullModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Render(fmt.Sprintf("Pulling %s...", m.repo.Name))

	logSection := ""
	if len(m.logs) > 0 {
		logSection = "\n" + strings.Join(m.logs, "\n")
	}

	gitBox := ""
	if len(m.gitBox.output) > 0 {
		gitBox = "\n" + m.gitBox.View()
	}

	if m.done {
		return fmt.Sprintf("%s%s\n%s\n", title, logSection, m.status)
	}

	statusLine := fmt.Sprintf("\n%s %s",
		m.spinner.View(),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(m.status))

	return fmt.Sprintf("%s%s%s%s", title, logSection, statusLine, gitBox)
}

func (m *pullModel) tryPull(url string) tea.Cmd {
	return func() tea.Msg {
		err := m.rs.pullRepoFromURLWithOutput(m.ctx, url, m.repo, m.updateRepoFromToml, m.writer)
		return tryPullMsg{err}
	}
}
