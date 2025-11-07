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

package gitmanager

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/storage/memory"
)

func getRefs(r *git.Repository) (memory.ReferenceStorage, error) {
	var zero memory.ReferenceStorage

	remote, err := r.Remote(git.DefaultRemoteName)
	if err != nil {
		return zero, err
	}

	endpoint, err := transport.NewEndpoint(remote.Config().URLs[0])
	if err != nil {
		return zero, err
	}

	gitClient, err := client.NewClient(endpoint)
	if err != nil {
		return zero, err
	}

	session, err := gitClient.NewUploadPackSession(endpoint, nil)
	if err != nil {
		return zero, err
	}

	info, err := session.AdvertisedReferences()
	if err != nil {
		return zero, err
	}

	refs, err := info.AllReferences()
	if err != nil {
		return zero, err
	}

	return refs, nil
}

func getHeadReference(r *git.Repository) (plumbing.ReferenceName, error) {
	refs, err := getRefs(r)
	if err != nil {
		return "", err
	}

	return refs["HEAD"].Target(), nil
}
