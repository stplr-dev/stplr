// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package list

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/pkg/types"
)

type ReposProvier interface {
	Repos() []types.Repo
}

type useCase struct {
	cfg    ReposProvier
	stdout io.Writer
}

func New(cfg ReposProvier) *useCase {
	return &useCase{cfg, os.Stdout}
}

type Options struct {
	Format string
	Json   bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	repos := u.cfg.Repos()
	if opts.Json {
		err := json.NewEncoder(u.stdout).Encode(repos)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error encoding repos to JSON"))
		}
		return nil
	}

	var tmpl *template.Template
	var err error
	format := opts.Format
	if format == "" {
		format = fmt.Sprintf(`%s: {{.Name}}
%s: {{.URL}}{{if .Ref}}
%s: {{.Ref}}{{end}}{{if .Mirrors}}
%s: {{range $i, $m := .Mirrors}}
  - {{$m}}{{end}}{{end}}{{if .ReportUrl}}
%s: {{.ReportUrl}}{{end}}

`, gotext.Get("Name"), gotext.Get("URL"), gotext.Get("Ref"), gotext.Get("Mirrors"), gotext.Get("Report"))
	}
	tmpl, err = template.New("format").Parse(format)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error parsing format template"))
	}

	for _, repo := range repos {
		err = tmpl.Execute(u.stdout, repo)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error executing template"))
		}
	}
	return nil
}
