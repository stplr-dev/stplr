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

package logger

import (
	"log/slog"
	"os"
)

var Level = new(slog.LevelVar)

func SetupDefault() {
	Level.Set(slog.LevelInfo)
	slogLogger := slog.New(NewJournalHandler(Level))
	slog.SetDefault(slogLogger)
}

func SetupForGoPlugin() {
	slogLogger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: Level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "@timestamp"
			case slog.MessageKey:
				a.Key = "@message"
			case slog.LevelKey:
				a.Key = "@level"
			}
			return a
		},
	}))
	slog.SetDefault(slogLogger)
}
