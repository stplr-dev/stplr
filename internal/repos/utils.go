// ALR - Any Linux Repository
// Copyright (C) 2025 The ALR Authors
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

package repos

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/alrsh"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/distro"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/types"
)

// isValid makes sure the path of the file being updated is valid.
// It checks to make sure the file is not within a nested directory
// and that it is called alr.sh.
func isValid(from, to diff.File) bool {
	var path string
	if from != nil {
		path = from.Path()
	}
	if to != nil {
		path = to.Path()
	}

	match, _ := filepath.Match("*/*.sh", path)
	return match
}

func parseScript(
	ctx context.Context,
	repo types.Repo,
	syntaxParser *syntax.Parser,
	runner *interp.Runner,
	r io.ReadCloser,
) ([]*alrsh.Package, error) {
	f, err := alrsh.ReadFromIOReader(r, "/tmp")
	if err != nil {
		return nil, err
	}
	_, dbPkgs, err := f.ParseBuildVars(ctx, &distro.OSRelease{}, []string{})
	if err != nil {
		return nil, err
	}
	for _, pkg := range dbPkgs {
		pkg.Repository = repo.Name
	}
	return dbPkgs, nil
}

func getHeadReference(r *git.Repository) (plumbing.ReferenceName, error) {
	remote, err := r.Remote(git.DefaultRemoteName)
	if err != nil {
		return "", err
	}

	endpoint, err := transport.NewEndpoint(remote.Config().URLs[0])
	if err != nil {
		return "", err
	}

	gitClient, err := client.NewClient(endpoint)
	if err != nil {
		return "", err
	}

	session, err := gitClient.NewUploadPackSession(endpoint, nil)
	if err != nil {
		return "", err
	}

	info, err := session.AdvertisedReferences()
	if err != nil {
		return "", err
	}

	refs, err := info.AllReferences()
	if err != nil {
		return "", err
	}

	return refs["HEAD"].Target(), nil
}

func resolveHash(r *git.Repository, ref string) (*plumbing.Hash, error) {
	var err error

	if ref == "" {
		reference, err := getHeadReference(r)
		if err != nil {
			return nil, fmt.Errorf("failed to get head reference %w", err)
		}
		ref = reference.Short()
	}

	hsh, err := r.ResolveRevision(git.DefaultRemoteName + "/" + plumbing.Revision(ref))
	if err != nil {
		hsh, err = r.ResolveRevision(plumbing.Revision(ref))
		if err != nil {
			return nil, err
		}
	}

	return hsh, nil
}
