// ALR - Any Linux Repository
// Copyright (C) 2025 Евгений Храмов
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

package dl_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitea.plemya-x.ru/Plemya-x/ALR/internal/config"
	"gitea.plemya-x.ru/Plemya-x/ALR/internal/dl"
	"gitea.plemya-x.ru/Plemya-x/ALR/internal/dlcache"
)

type TestALRConfig struct{}

func (c *TestALRConfig) GetPaths(ctx context.Context) *config.Paths {
	return &config.Paths{
		CacheDir: "/tmp",
	}
}

func TestDownloadWithoutCache(t *testing.T) {
	type testCase struct {
		name     string
		path     string
		expected func(*testing.T, error, string)
	}

	prepareServer := func() *httptest.Server {
		// URL вашего Git-сервера
		gitServerURL, err := url.Parse("https://gitea.plemya-x.ru")
		if err != nil {
			log.Fatalf("Failed to parse git server URL: %v", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(gitServerURL)

		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/file-downloader/file":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, World!"))
			case strings.HasPrefix(r.URL.Path, "/git-downloader/git"):
				r.URL.Host = gitServerURL.Host
				r.URL.Scheme = gitServerURL.Scheme
				r.Host = gitServerURL.Host
				r.URL.Path, _ = strings.CutPrefix(r.URL.Path, "/git-downloader/git")

				proxy.ServeHTTP(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	}

	for _, tc := range []testCase{
		{
			name: "simple file download",
			path: "%s/file-downloader/file",
			expected: func(t *testing.T, err error, tmpdir string) {
				assert.NoError(t, err)

				_, err = os.Stat(path.Join(tmpdir, "file"))
				assert.NoError(t, err)
			},
		},
		{
			name: "git download",
			path: "git+%s/git-downloader/git/Plemya-x/xpamych-alr-repo",
			expected: func(t *testing.T, err error, tmpdir string) {
				assert.NoError(t, err)

				_, err = os.Stat(path.Join(tmpdir, "alr-repo.toml"))
				assert.NoError(t, err)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := prepareServer()
			defer server.Close()

			tmpdir, err := os.MkdirTemp("", "test-download")
			assert.NoError(t, err)
			defer os.RemoveAll(tmpdir)

			opts := dl.Options{
				CacheDisabled: true,
				URL:           fmt.Sprintf(tc.path, server.URL),
				Destination:   tmpdir,
			}

			err = dl.Download(context.Background(), opts)

			tc.expected(t, err, tmpdir)
		})
	}
}

func TestDownloadFileWithCache(t *testing.T) {
	type testCase struct {
		name string
	}

	for _, tc := range []testCase{
		{
			name: "simple download",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/file":
					called += 1
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Hello, World!"))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			tmpdir, err := os.MkdirTemp("", "test-download")
			assert.NoError(t, err)
			defer os.RemoveAll(tmpdir)

			cfg := &TestALRConfig{}

			opts := dl.Options{
				CacheDisabled: false,
				URL:           server.URL + "/file",
				Destination:   tmpdir,
				DlCache:       dlcache.New(cfg),
			}

			outputFile := path.Join(tmpdir, "file")

			err = dl.Download(context.Background(), opts)
			assert.NoError(t, err)
			_, err = os.Stat(outputFile)
			assert.NoError(t, err)

			err = os.Remove(outputFile)
			assert.NoError(t, err)

			err = dl.Download(context.Background(), opts)
			assert.NoError(t, err)
			assert.Equal(t, 1, called)
		})
	}
}
