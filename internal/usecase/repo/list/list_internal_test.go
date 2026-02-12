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
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"go.stplr.dev/stplr/pkg/types"
)

func TestListUseCaseRun(t *testing.T) {
	mustJsonMarshal := func(v any) []byte {
		b, err := json.Marshal(v)
		assert.NoError(t, err)
		return b
	}

	tests := []struct {
		name           string
		repos          []types.Repo
		format         string
		json           bool
		expectedOutput string
		expectError    bool
	}{
		{
			name: "Default format with multiple repos",
			repos: []types.Repo{
				{Name: "repo1", URL: "http://repo1.com"},
				{Name: "repo2", URL: "http://repo2.com"},
			},
			format:         "",
			json:           false,
			expectedOutput: "Name: repo1\nURL: http://repo1.com\n\nName: repo2\nURL: http://repo2.com\n\n",
			expectError:    false,
		},
		{
			name: "Custom format with multiple repos",
			repos: []types.Repo{
				{Name: "repo1", URL: "http://repo1.com"},
				{Name: "repo2", URL: "http://repo2.com"},
			},
			format:         "{{.Name}}: {{.URL}}\n",
			json:           false,
			expectedOutput: "repo1: http://repo1.com\nrepo2: http://repo2.com\n",
			expectError:    false,
		},
		{
			name:           "Empty repos list",
			repos:          []types.Repo{},
			format:         "",
			json:           false,
			expectedOutput: "",
			expectError:    false,
		},
		{
			name: "Invalid template format",
			repos: []types.Repo{
				{Name: "repo1", URL: "http://repo1.com"},
			},
			format:         "{{.InvalidField}}\n",
			json:           false,
			expectedOutput: "",
			expectError:    true,
		},
		{
			name: "JSON format with multiple repos",
			repos: []types.Repo{
				{Name: "repo1", URL: "http://repo1.com"},
				{Name: "repo2", URL: "http://repo2.com"},
			},
			format: "",
			json:   true,
			expectedOutput: string(mustJsonMarshal([]types.Repo{
				{Name: "repo1", URL: "http://repo1.com"},
				{Name: "repo2", URL: "http://repo2.com"},
			})),
			expectError: false,
		},
		{
			name: "Default format with additional fields",
			repos: []types.Repo{
				{
					Name:      "repo1",
					URL:       "http://repo1.com",
					Ref:       "main",
					Mirrors:   []string{"http://mirror1.com", "http://mirror2.com"},
					ReportUrl: "http://report.com",
					Title:     "Repo1",
					Summary:   "Short summary",
					Description: `Long multiline
description`,
					Homepage: "http://repo1.com",
					Icon:     "http://repo1.com/icon.svg",
				},
			},
			format: "",
			json:   false,
			expectedOutput: `Name: repo1
Title: Repo1
Summary: Short summary
Description:
  Long multiline
  description
Homepage: http://repo1.com
Icon: http://repo1.com/icon.svg
URL: http://repo1.com
Ref: main
Mirrors: 
  - http://mirror1.com
  - http://mirror2.com
Report: http://report.com

`,
			expectError: false,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockReposProvier(ctrl)
			mockProvider.EXPECT().Repos().Return(tt.repos)

			useCase := New(mockProvider)

			// Capture output
			var buf bytes.Buffer
			useCase.stdout = &buf

			opts := Options{Format: tt.format, Json: tt.json}
			err := useCase.Run(context.Background(), opts)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.json {
					// For JSON output, we need to compare the parsed JSON to handle formatting differences
					var expected, actual []types.Repo
					json.Unmarshal([]byte(tt.expectedOutput), &expected)
					json.Unmarshal(buf.Bytes(), &actual)
					assert.Equal(t, expected, actual)
				} else {
					assert.Equal(t, tt.expectedOutput, buf.String())
				}
			}
		})
	}
}
