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

package dl

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leonelquinteros/gotext"
)

type model struct {
	progress   progress.Model
	spinner    spinner.Model
	percent    float64
	speed      float64
	done       bool
	useSpinner bool
	filename   string

	total      int64
	downloaded int64
	elapsed    time.Duration
	remaining  time.Duration

	width int
}

func (m model) Init() tea.Cmd {
	if m.useSpinner {
		return m.spinner.Tick
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case progressUpdate:
		m.percent = msg.percent
		m.speed = msg.speed
		m.downloaded = msg.downloaded
		m.total = msg.total
		m.elapsed = time.Duration(msg.elapsed) * time.Second
		m.remaining = time.Duration(msg.remaining) * time.Second
		if m.percent >= 1.0 {
			m.done = true
			return m, tea.Quit
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case progress.FrameMsg:
		if !m.useSpinner {
			progressModel, cmd := m.progress.Update(msg)
			m.progress = progressModel.(progress.Model)
			return m, cmd
		}
	case spinner.TickMsg:
		if m.useSpinner {
			spinnerModel, cmd := m.spinner.Update(msg)
			m.spinner = spinnerModel
			return m, cmd
		}
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.done {
		return gotext.Get("%s: done!\n", m.filename)
	}
	if m.useSpinner {
		return gotext.Get(
			"%s %s downloading at %s/s\n",
			m.filename,
			m.spinner.View(),
			prettyByteSize(int64(m.speed)),
		)
	}

	leftPart := m.filename

	rightPart := fmt.Sprintf("%.2f%% (%s/%s, %s/s) [%v:%v]\n", m.percent*100,
		prettyByteSize(m.downloaded),
		prettyByteSize(m.total),
		prettyByteSize(int64(m.speed)),
		m.elapsed,
		m.remaining,
	)

	m.progress.Width = m.width - len(leftPart) - len(rightPart) - 6
	bar := m.progress.ViewAs(m.percent)
	return fmt.Sprintf(
		"%s %s %s",
		leftPart,
		bar,
		rightPart,
	)
}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

type progressUpdate struct {
	percent float64
	speed   float64
	total   int64

	downloaded int64
	elapsed    float64
	remaining  float64
}

type ProgressWriter struct {
	baseWriter   io.WriteCloser
	total        int64
	downloaded   int64
	startTime    time.Time
	onProgress   func(progressUpdate)
	lastReported time.Time
	doneChan     chan struct{}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.baseWriter.Write(p)
	if err != nil {
		return n, err
	}

	pw.downloaded += int64(n)
	now := time.Now()
	elapsed := now.Sub(pw.startTime).Seconds()
	speed := float64(pw.downloaded) / elapsed
	var remaining, percent float64
	if pw.total > 0 {
		remaining = (float64(pw.total) - float64(pw.downloaded)) / speed
		percent = float64(pw.downloaded) / float64(pw.total)
	}

	if now.Sub(pw.lastReported) > 100*time.Millisecond {
		pw.onProgress(progressUpdate{
			percent:    percent,
			speed:      speed,
			total:      pw.total,
			downloaded: pw.downloaded,
			elapsed:    elapsed,
			remaining:  remaining,
		})
		pw.lastReported = now
	}

	return n, nil
}

func (pw *ProgressWriter) Close() error {
	pw.onProgress(progressUpdate{
		percent:    1,
		speed:      0,
		downloaded: pw.downloaded,
	})
	<-pw.doneChan
	return nil
}

func NewProgressWriter(base io.WriteCloser, max int64, filename string, out io.Writer) *ProgressWriter {
	var m *model
	if max == -1 {
		m = &model{
			spinner:    spinner.New(),
			useSpinner: true,
			filename:   filename,
		}
		m.spinner.Spinner = spinner.Dot
	} else {
		m = &model{
			progress: progress.New(
				progress.WithDefaultGradient(),
				progress.WithoutPercentage(),
			),
			useSpinner: false,
			filename:   filename,
		}
	}

	p := tea.NewProgram(m,
		tea.WithInput(nil),
		tea.WithOutput(out),
	)

	pw := &ProgressWriter{
		baseWriter: base,
		total:      max,
		startTime:  time.Now(),
		doneChan:   make(chan struct{}),
		onProgress: func(update progressUpdate) {
			p.Send(update)
		},
	}

	go func() {
		defer close(pw.doneChan)
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running progress writer: %v\n", err)
			os.Exit(1)
		}
	}()

	return pw
}
