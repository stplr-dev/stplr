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

package remove

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/pkg/types"
)

type Repos interface {
	GetRepo(name string) (types.Repo, error)
	SetOverride(ctx context.Context, name string, o types.RepoOverride, pull bool) error
}

type useCase struct {
	r Repos
}

func New(r Repos) *useCase {
	return &useCase{r}
}

type Options struct {
	Name          string
	URL           string
	IgnoreMissing bool
	PartialMatch  bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	repo, err := u.r.GetRepo(opts.Name)
	if err != nil {
		if opts.IgnoreMissing {
			return nil
		}
		return err
	}

	original := repo.Mirrors
	filtered := slices.DeleteFunc(slices.Clone(original), func(mirror string) bool {
		if opts.PartialMatch {
			return strings.Contains(mirror, opts.URL)
		}
		return mirror == opts.URL
	})

	removed := len(original) - len(filtered)
	if removed == 0 {
		if opts.IgnoreMissing {
			return nil
		}
		if opts.PartialMatch {
			return errors.NewI18nError(gotext.Get("No mirrors containing \"%s\" found in repo \"%s\"", opts.URL, opts.Name))
		}
		return errors.NewI18nError(gotext.Get("URL \"%s\" does not exist in repo \"%s\"", opts.URL, opts.Name))
	}

	if err := u.r.SetOverride(ctx, opts.Name, types.RepoOverride{Mirrors: filtered}, false); err != nil {
		return err
	}

	fmt.Println(gotext.Get("Removed %d mirrors from repo \"%s\"\n", removed, opts.Name))
	return nil
}
