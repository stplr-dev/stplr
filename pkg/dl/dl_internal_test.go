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

package dl

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDownloadWithTimeouts(t *testing.T) {
	type testCase struct {
		name        string
		setupServer func() *httptest.Server
		expected    func(*testing.T, error)
	}

	commonTimeoutDuration = 1 * time.Second

	for _, tc := range []testCase{
		{
			name: "response header timeout",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate slow header response (longer than 30 seconds)
					time.Sleep(commonTimeoutDuration + 1*time.Second)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("too late"))
				}))
			},
			expected: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "connection timeout",
			setupServer: func() *httptest.Server {
				// Create a server that accepts connections but never responds
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				assert.NoError(t, err)

				server := &httptest.Server{
					Listener: listener,
					Config: &http.Server{
						Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							time.Sleep(commonTimeoutDuration + 5*time.Second)
						}),
					},
				}
				server.Start()
				return server
			},
			expected: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "successful download within timeout",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Respond quickly (well within timeout)
					time.Sleep(100 * time.Millisecond)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				}))
			},
			expected: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "idle connection reuse",
			setupServer: func() *httptest.Server {
				connectionCount := 0
				var mu sync.Mutex

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					connectionCount++
					count := connectionCount
					mu.Unlock()

					w.Header().Set("X-Connection-Count", fmt.Sprintf("%d", count))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("data"))
				}))
			},
			expected: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer server.Close()

			tmpdir, err := os.MkdirTemp("", "test-download-timeout")
			assert.NoError(t, err)
			defer os.RemoveAll(tmpdir)

			opts := Options{
				CacheDisabled: true,
				URL:           server.URL + "/file",
				Destination:   tmpdir,
			}

			ctx, cancel := context.WithTimeout(context.Background(), commonTimeoutDuration+10*time.Second)
			defer cancel()

			err = Download(ctx, opts)
			tc.expected(t, err)
		})
	}
}
