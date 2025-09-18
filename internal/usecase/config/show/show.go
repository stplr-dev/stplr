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

package show

import (
	"context"
	"fmt"
	"os"
)

type ConfigProvider interface {
	ToYAML() (string, error)
}

type useCase struct {
	config ConfigProvider

	stdout *os.File
}

func New(config ConfigProvider) *useCase {
	return &useCase{
		config: config,
		stdout: os.Stdout,
	}
}

func (u *useCase) Run(ctx context.Context) error {
	content, err := u.config.ToYAML()
	if err != nil {
		return err
	}
	fmt.Fprintln(u.stdout, content)
	return nil
}
