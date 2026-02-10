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

package updater_test

import (
	context "context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"go.stplr.dev/stplr/internal/service/updater"
	"go.stplr.dev/stplr/pkg/distro"
	staplerfile "go.stplr.dev/stplr/pkg/staplerfile"
)

func TestUpdaterCheckForUpdates(t *testing.T) {
	tests := []struct {
		name              string
		installedPackages map[string]string
		ignorePatterns    []string
		searchResults     map[string][]staplerfile.Package
		searchErrors      map[string]error
		listInstalledErr  error
		expected          []updater.UpdateInfo
		expectError       bool
	}{
		{
			name: "no updates available",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
			},
			ignorePatterns: []string{},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.0.0",
					},
				},
			},
			expected:    []updater.UpdateInfo{},
			expectError: false,
		},
		{
			name: "update available",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
			},
			ignorePatterns: []string{},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "2.0.0",
					},
				},
			},
			expected: []updater.UpdateInfo{
				{
					Package: &staplerfile.Package{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "2.0.0",
					},
					FromVersion: "1.0.0",
					ToVersion:   "2.0.0",
				},
			},
			expectError: false,
		},
		{
			name: "update available with release",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0-1",
			},
			ignorePatterns: []string{},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.0.0",
						Release:    2,
					},
				},
			},
			expected: []updater.UpdateInfo{
				{
					Package: &staplerfile.Package{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.0.0",
						Release:    2,
					},
					FromVersion: "1.0.0-1",
					ToVersion:   "1.0.0-2",
				},
			},
			expectError: false,
		},
		{
			name: "update available with epoch",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1:1.0.0-1",
			},
			ignorePatterns: []string{},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.0.0",
						Release:    2,
						Epoch:      1,
					},
				},
			},
			expected: []updater.UpdateInfo{
				{
					Package: &staplerfile.Package{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.0.0",
						Release:    2,
						Epoch:      1,
					},
					FromVersion: "1:1.0.0-1",
					ToVersion:   "1:1.0.0-2",
				},
			},
			expectError: false,
		},
		{
			name: "package ignored by pattern",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
			},
			ignorePatterns: []string{"repo/pkg1"},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "2.0.0",
					},
				},
			},
			expected:    []updater.UpdateInfo{},
			expectError: false,
		},
		{
			name: "package not found in search",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
			},
			ignorePatterns: []string{},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {},
			},
			expected:    []updater.UpdateInfo{},
			expectError: false,
		},
		{
			name:              "list installed returns error",
			installedPackages: map[string]string{},
			listInstalledErr:  errors.New("list error"),
			expected:          nil,
			expectError:       true,
		},
		{
			name: "search returns error",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
			},
			ignorePatterns: []string{},
			searchErrors: map[string]error{
				"pkg1+stplr-repo": errors.New("search error"),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "multiple packages with updates",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
				"pkg2+stplr-repo": "2.0.0",
			},
			ignorePatterns: []string{},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.5.0",
					},
				},
				"pkg2+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg2",
						Version:    "3.0.0",
					},
				},
			},
			expected: []updater.UpdateInfo{
				{
					Package: &staplerfile.Package{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "1.5.0",
					},
					FromVersion: "1.0.0",
					ToVersion:   "1.5.0",
				},
				{
					Package: &staplerfile.Package{
						Repository: "repo",
						Name:       "pkg2",
						Version:    "3.0.0",
					},
					FromVersion: "2.0.0",
					ToVersion:   "3.0.0",
				},
			},
			expectError: false,
		},
		{
			name: "package ignored by glob pattern",
			installedPackages: map[string]string{
				"pkg1+stplr-repo": "1.0.0",
				"pkg2+stplr-foo":  "1.0.0",
			},
			ignorePatterns: []string{"repo/*"},
			searchResults: map[string][]staplerfile.Package{
				"pkg1+stplr-repo": {
					{
						Repository: "repo",
						Name:       "pkg1",
						Version:    "2.0.0",
					},
				},
				"pkg2+stplr-foo": {
					{
						Repository: "foo",
						Name:       "pkg2",
						Version:    "2.0.0",
					},
				},
			},
			expected: []updater.UpdateInfo{
				{
					Package: &staplerfile.Package{
						Repository: "foo",
						Name:       "pkg2",
						Version:    "2.0.0",
					},
					FromVersion: "1.0.0",
					ToVersion:   "2.0.0",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := updater.NewMockIgnoreUpdatesProvider(ctrl)
			mgr := updater.NewMockManager(ctrl)
			searcher := updater.NewMockSearcher(ctrl)
			info := &distro.OSRelease{}

			updater := updater.New(cfg, mgr, info, searcher)

			mgr.EXPECT().
				ListInstalled(nil).
				Return(tt.installedPackages, tt.listInstalledErr).
				Times(1)

			if tt.listInstalledErr == nil {
				cfg.EXPECT().
					IgnorePkgUpdates().
					Return(tt.ignorePatterns).
					AnyTimes()

				for pkgName := range tt.installedPackages {
					if tt.searchResults != nil {
						if results, ok := tt.searchResults[pkgName]; ok {
							searcher.EXPECT().
								Search(gomock.Any(), gomock.Any()).
								Return(results, nil).
								MaxTimes(1)
						}
					}
					if tt.searchErrors != nil {
						if err, ok := tt.searchErrors[pkgName]; ok {
							searcher.EXPECT().
								Search(gomock.Any(), gomock.Any()).
								Return(nil, err).
								MaxTimes(1)
						}
					}
				}
			}

			result, err := updater.CheckForUpdates(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, len(tt.expected))

				for i, expected := range tt.expected {
					if i < len(result) {
						assert.Equal(t, expected.FromVersion, result[i].FromVersion)
						assert.Equal(t, expected.ToVersion, result[i].ToVersion)
						assert.Equal(t, expected.Package.Repository, result[i].Package.Repository)
						assert.Equal(t, expected.Package.Name, result[i].Package.Name)
						assert.Equal(t, expected.Package.Version, result[i].Package.Version)
					}
				}
			}
		})
	}
}
