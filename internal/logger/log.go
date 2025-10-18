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

package logger

import (
	"context"
	"log/slog"
	"os"

	chLog "github.com/charmbracelet/log"
)

type Logger struct {
	l *chLog.Logger
}

func setupLogger() *chLog.Logger {
	// styles := chLog.DefaultStyles()
	logger := chLog.New(os.Stderr)
	// styles.Levels[chLog.InfoLevel] = lipgloss.NewStyle().
	// 	SetString("-->").
	// 	Foreground(lipgloss.Color("35"))
	// styles.Levels[chLog.ErrorLevel] = lipgloss.NewStyle().
	// 	SetString(gotext.Get("ERROR")).
	// 	Padding(0, 1, 0, 1).
	// 	Background(lipgloss.Color("204")).
	// 	Foreground(lipgloss.Color("0"))
	// logger.SetStyles(styles)
	return logger
}

func New() *Logger {
	return &Logger{
		l: setupLogger(),
	}
}

func slogLevelToLog(level slog.Level) chLog.Level {
	switch level {
	case slog.LevelDebug:
		return chLog.DebugLevel
	case slog.LevelInfo:
		return chLog.InfoLevel
	case slog.LevelWarn:
		return chLog.WarnLevel
	case slog.LevelError:
		return chLog.ErrorLevel
	}
	return chLog.FatalLevel
}

func (l *Logger) SetLevel(level slog.Level) {
	l.l.SetLevel(slogLevelToLog(level))
}

func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.l.Enabled(ctx, level)
}

func (l *Logger) Handle(ctx context.Context, rec slog.Record) error {
	return l.l.Handle(ctx, rec)
}

func (l *Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	sl := *l
	sl.l = l.l.WithAttrs(attrs).(*chLog.Logger)
	return &sl
}

func (l *Logger) WithGroup(name string) slog.Handler {
	sl := *l
	sl.l = l.l.WithGroup(name).(*chLog.Logger)
	return &sl
}

var logger *Logger

func SetupDefault() *Logger {
	logger = New()
	slogLogger := slog.New(logger)
	slog.SetDefault(slogLogger)
	return logger
}

func SetupForGoPlugin() {
	logger.l.SetFormatter(chLog.JSONFormatter)
	chLog.TimestampKey = "@timestamp"
	chLog.MessageKey = "@message"
	chLog.LevelKey = "@level"
}

func GetLogger() *Logger {
	return logger
}
