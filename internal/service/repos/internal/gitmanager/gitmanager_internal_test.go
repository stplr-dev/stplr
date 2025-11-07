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
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := &GitManager{}
	ctx := context.Background()
	opts := gm.defaultFetchOptions()
	mockRepo := NewMockgitRepository(ctrl)

	t.Run("Successful fetch", func(t *testing.T) {
		mockRepo.EXPECT().FetchContext(ctx, opts).Return(nil)

		err := gm.fetch(ctx, mockRepo, opts)
		assert.NoError(t, err)
	})

	t.Run("Already up to date", func(t *testing.T) {
		mockRepo.EXPECT().FetchContext(ctx, opts).Return(git.NoErrAlreadyUpToDate)

		err := gm.fetch(ctx, mockRepo, opts)
		assert.NoError(t, err)
	})

	t.Run("Fetch error", func(t *testing.T) {
		fetchErr := errors.New("fetch failed")
		mockRepo.EXPECT().FetchContext(ctx, opts).Return(fetchErr)

		err := gm.fetch(ctx, mockRepo, opts)
		assert.ErrorIs(t, err, fetchErr)
	})
}

func TestFetchRepoByRef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := &GitManager{}
	ctx := context.Background()
	ref := "refs/heads/main"
	tempTag := "stplr-temp-tag-refs/heads/main"
	mockRepo := NewMockgitRepository(ctrl)

	t.Run("Successful fetch and tag deletion", func(t *testing.T) {
		opts := gm.defaultFetchOptions()
		opts.RefSpecs = append(opts.RefSpecs, config.RefSpec(fmt.Sprintf("%s:refs/tags/%s", ref, tempTag)))
		mockRepo.EXPECT().FetchContext(ctx, opts).Return(nil)
		mockRepo.EXPECT().DeleteTag(tempTag).Return(nil)

		err := gm.fetchRepoByRef(ctx, mockRepo, ref)
		assert.NoError(t, err)
	})

	t.Run("Invalid refspec", func(t *testing.T) {
		invalidRef := "invalid:ref"
		err := gm.fetchRepoByRef(ctx, mockRepo, invalidRef)
		assert.Error(t, err)
	})

	t.Run("Tag deletion failure", func(t *testing.T) {
		opts := gm.defaultFetchOptions()
		opts.RefSpecs = append(opts.RefSpecs, config.RefSpec(fmt.Sprintf("%s:refs/tags/%s", ref, tempTag)))
		mockRepo.EXPECT().FetchContext(ctx, opts).Return(nil)
		mockRepo.EXPECT().DeleteTag(tempTag).Return(errors.New("delete tag failed"))

		err := gm.fetchRepoByRef(ctx, mockRepo, ref)
		assert.NoError(t, err) // Error in defer is logged, not returned
	})
}
