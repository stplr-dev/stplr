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

package output

import "context"

type Output interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type contextKey struct{}

var outputKey = contextKey{}

func WithOutput(ctx context.Context, out Output) context.Context {
	return context.WithValue(ctx, outputKey, out)
}

func FromContext(ctx context.Context) Output {
	if v := ctx.Value(outputKey); v != nil {
		if out, ok := v.(Output); ok {
			return out
		}
	}
	return NewConsoleOutput()
}
