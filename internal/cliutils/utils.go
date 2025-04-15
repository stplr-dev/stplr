// ALR - Any Linux Repository
// Copyright (C) 2025 Евгений Храмов
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

package cliutils

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"
)

type BashCompleteWithErrorFunc func(c *cli.Context) error

func BashCompleteWithError(f BashCompleteWithErrorFunc) cli.BashCompleteFunc {
	return func(c *cli.Context) { HandleExitCoder(f(c)) }
}

func HandleExitCoder(err error) {
	if err == nil {
		return
	}

	if exitErr, ok := err.(cli.ExitCoder); ok {
		if err.Error() != "" {
			if _, ok := exitErr.(cli.ErrorFormatter); ok {
				slog.Error(fmt.Sprintf("%+v\n", err))
			} else {
				slog.Error(err.Error())
			}
		}
		cli.OsExiter(exitErr.ExitCode())
		return
	}
}

func FormatCliExit(msg string, err error) cli.ExitCoder {
	return FormatCliExitWithCode(msg, err, 1)
}

func FormatCliExitWithCode(msg string, err error, exitCode int) cli.ExitCoder {
	if err == nil {
		return cli.Exit(errors.New(msg), exitCode)
	}
	return cli.Exit(fmt.Errorf("%s: %w", msg, err), exitCode)
}
