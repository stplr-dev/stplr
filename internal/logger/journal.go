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
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/coreos/go-systemd/v22/journal"
)

type JournalHandler struct {
	level *slog.LevelVar
}

func NewJournalHandler(lv *slog.LevelVar) *JournalHandler {
	return &JournalHandler{level: lv}
}

func (h *JournalHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

func (h *JournalHandler) Handle(_ context.Context, r slog.Record) error {
	return journal.Send(formatSingleLine(r), toPriority(r.Level), nil)
}

func (h *JournalHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// journald does not have native scoped attributes, so just flatten them
	return &JournalHandler{level: h.level}
}

func (h *JournalHandler) WithGroup(name string) slog.Handler {
	// journald does not support groups; ignore
	return h
}

func (h *JournalHandler) LevelVar() *slog.LevelVar {
	return h.level
}

func toPriority(lvl slog.Level) journal.Priority {
	switch {
	case lvl <= slog.LevelDebug:
		return journal.PriDebug
	case lvl <= slog.LevelInfo:
		return journal.PriInfo
	case lvl <= slog.LevelWarn:
		return journal.PriWarning
	default:
		return journal.PriErr
	}
}

func formatSingleLine(r slog.Record) string {
	var b strings.Builder

	b.WriteString(r.Message)

	r.Attrs(func(a slog.Attr) bool {
		b.WriteByte(' ')
		b.WriteString(formatKeyValue(a))
		return true
	})

	return b.String()
}

func formatKeyValue(a slog.Attr) string {
	val := a.Value
	switch val.Kind() {
	case slog.KindString:
		return fmt.Sprintf("%s=%s", a.Key, val.String())
	case slog.KindInt64:
		return fmt.Sprintf("%s=%d", a.Key, val.Int64())
	case slog.KindBool:
		return fmt.Sprintf("%s=%t", a.Key, val.Bool())
	case slog.KindAny:
		return fmt.Sprintf("%s=%v", a.Key, val.Any())
	default:
		return fmt.Sprintf("%s=%v", a.Key, val)
	}
}
