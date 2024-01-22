/*
 * LURE - Linux User REpository
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package handlers

import (
	"context"
	"fmt"
	"time"

	"mvdan.cc/sh/v3/interp"
)

func InsufficientArgsError(cmd string, exp, got int) error {
	argsWord := "arguments"
	if exp == 1 {
		argsWord = "argument"
	}

	return fmt.Errorf("%s: command requires at least %d %s, got %d", cmd, exp, argsWord, got)
}

type ExecFunc func(hc interp.HandlerContext, name string, args []string) error

type ExecFuncs map[string]ExecFunc

// ExecHandler returns a new ExecHandlerFunc that falls back to fallback
// if the command cannot be found in the map. If fallback is nil, the default
// handler is used.
func (ef ExecFuncs) ExecHandler(fallback interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		name := args[0]

		if fn, ok := ef[name]; ok {
			hctx := interp.HandlerCtx(ctx)
			if len(args) > 1 {
				return fn(hctx, args[0], args[1:])
			} else {
				return fn(hctx, args[0], nil)
			}
		}

		if fallback == nil {
			fallback = interp.DefaultExecHandler(2 * time.Second)
		}
		return fallback(ctx, args)
	}
}
