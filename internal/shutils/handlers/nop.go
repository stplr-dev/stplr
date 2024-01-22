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
	"io"
	"os"
)

func NopReadDir(context.Context, string) ([]os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func NopStat(context.Context, string, bool) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func NopExec(context.Context, []string) error {
	return nil
}

func NopOpen(context.Context, string, int, os.FileMode) (io.ReadWriteCloser, error) {
	return NopRWC{}, nil
}

type NopRWC struct{}

func (NopRWC) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (NopRWC) Write(b []byte) (int, error) {
	return len(b), nil
}

func (NopRWC) Close() error {
	return nil
}
