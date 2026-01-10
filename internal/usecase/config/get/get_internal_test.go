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

package get

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/common"
	"go.stplr.dev/stplr/pkg/types"
)

func TestAllAllowedKeysHandled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().RootCmd().Return("1")
	mockConfig.EXPECT().PagerStyle().Return("1")
	mockConfig.EXPECT().AutoPull().Return(true)
	mockConfig.EXPECT().LogLevel().Return("1")
	mockConfig.EXPECT().UseRootCmd().Return(true)
	mockConfig.EXPECT().IgnorePkgUpdates().Return([]string{})
	mockConfig.EXPECT().ForbidBuildCommand().Return(true)
	mockConfig.EXPECT().ForbidSkipInChecksums().Return(true)
	mockConfig.EXPECT().HideFirejailExcludeWarning().Return(true)
	mockConfig.EXPECT().FirejailExclude().Return([]string{})

	for _, key := range config.AllowedKeys() {
		useCase := New(mockConfig)
		buf := bytes.NewBuffer(nil)
		useCase.out = buf
		useCase.Run(t.Context(), key)
		assert.NotEmpty(t, buf)
	}
}

func TestLegacyKeysHandled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().Repos().Return([]types.Repo{})
	mockConfig.EXPECT().Repos().Return([]types.Repo{})

	for _, key := range []string{"repo", "repos"} {
		useCase := New(mockConfig)
		buf := bytes.NewBuffer(nil)
		useCase.out = buf
		useCase.Run(t.Context(), key)
		assert.NotEmpty(t, buf)
	}
}

func TestStringKeyOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().RootCmd().Return("test-cmd")

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	useCase.Run(t.Context(), common.ROOT_CMD)

	assert.Equal(t, "test-cmd\n", buf.String())
}

func TestBoolKeyOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().AutoPull().Return(true)

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	useCase.Run(t.Context(), common.AUTO_PULL)

	assert.Equal(t, "true\n", buf.String())
}

func TestListKeyOutputEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().IgnorePkgUpdates().Return([]string{})

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	useCase.Run(t.Context(), common.IGNORE_PKG_UPDATES)

	assert.Equal(t, "[]\n", buf.String())
}

func TestListKeyOutputWithValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().IgnorePkgUpdates().Return([]string{"pkg1", "pkg2", "pkg3"})

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	useCase.Run(t.Context(), common.IGNORE_PKG_UPDATES)

	assert.Equal(t, "pkg1, pkg2, pkg3\n", buf.String())
}

func TestReposKeyEmptyOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().Repos().Return([]types.Repo{})

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	useCase.Run(t.Context(), "repo")

	assert.Equal(t, "[]\n", buf.String())
}

func TestReposKeyWithValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)
	mockConfig.EXPECT().Repos().Return([]types.Repo{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo2", URL: "https://example.com/repo2"},
	})

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	err := useCase.Run(t.Context(), "repo")

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "repo1")
	assert.Contains(t, buf.String(), "repo2")
}

func TestUnknownKeyReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := NewMockConfigGetter(ctrl)

	useCase := New(mockConfig)
	buf := bytes.NewBuffer(nil)
	useCase.out = buf
	err := useCase.Run(t.Context(), "unknown_key")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}
