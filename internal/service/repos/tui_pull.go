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
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/plugins/shared"
	"go.stplr.dev/stplr/internal/service/repos/internal/puller"
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
	case notifyWriteMsg:

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

type pullModel struct {
	ctx  context.Context
	done bool

	repo types.Repo

	lastUrl string
	logs    []string
	status  string

	rs PullExecutor

	spinner spinner.Model
	gitBox  progressViewport

	writer             io.Writer
	updateRepoFromToml bool

	notifier *teaNotifier
	msgs     chan tea.Msg
}

func newPullModel(ctx context.Context, repo types.Repo, rs PullExecutor, updateRepoFromToml bool) *pullModel {
	pm := pullModel{}

	msgs := make(chan tea.Msg)
	notifier := &teaNotifier{parent: &pm, ch: msgs}

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

	pm.ctx = ctx
	pm.repo = repo
	pm.rs = rs
	pm.spinner = s
	pm.status = gotext.Get("Pull %s", urls[0])
	pm.logs = []string{}
	pm.gitBox = gitBox
	pm.updateRepoFromToml = updateRepoFromToml
	pm.notifier = notifier
	pm.msgs = msgs

	return &pm
}

func waitForNotifier(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch
		return msg
	}
}

type (
	doneMsg struct{}
	errMsg  struct {
		err error
	}
)

func (m *pullModel) pull() tea.Cmd {
	return func() tea.Msg {
		err := m.Pull()
		if err != nil {
			return errMsg{err}
		}
		return doneMsg{}
	}
}

func (m pullModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.pull(),
		waitForNotifier(m.msgs),
	)
}

func (m pullModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case notifyMsg:
		switch msg.Event {
		case puller.EventTryPull:
			i, _ := strconv.Atoi(msg.Data["i"])
			url := msg.Data["url"]
			if i == 0 {
				m.status = gotext.Get("Pull %s", url)
			} else {
				m.status = gotext.Get("Trying mirror %d: %s", i, url)
			}
			m.lastUrl = url
			return m, waitForNotifier(m.msgs)
		case puller.EventErrorPull:
			line := textErorrDarkerStyle.
				Render(gotext.Get("- Failed to pull from %s: %v", msg.Data["url"], strings.TrimSpace(msg.Data["err"])))
			m.logs = append(m.logs, line)
			return m, waitForNotifier(m.msgs)
		}
	case doneMsg:
		line := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render(gotext.Get("- Pulled from %s", m.lastUrl))
		m.logs = append(m.logs, line)
		m.status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).Render(gotext.Get("✔ Repository pulled successfully!"))
		m.done = true
		return m, tea.Quit
	case errMsg:
		m.done = true
		m.status = textErorrStyle.Render(gotext.Get("🞮 Failed to pull — the only source is unavailable"))
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
		Render(gotext.Get("Pulling %s...", m.repo.Name))

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

func (m *pullModel) Pull() error {
	newRepo, err := m.rs.Pull(m.ctx, m.repo, m.notifier)
	if err == nil {
		m.repo = newRepo
	}
	return err
}

type notifyMsg struct {
	Event shared.NotifyEvent
	Data  map[string]string
}

type notifyWriteMsg struct {
	Event shared.NotifyEvent
	Bytes []byte
}

type teaNotifier struct {
	parent *pullModel
	ch     chan tea.Msg
}

func (tn *teaNotifier) Notify(ctx context.Context, event shared.NotifyEvent, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case tn.ch <- notifyMsg{Event: event, Data: data}:
		return nil
	}
}

func (tn *teaNotifier) NotifyWrite(ctx context.Context, event shared.NotifyWriterEvent, p []byte) (n int, err error) {
	return tn.parent.writer.Write(p)
}
