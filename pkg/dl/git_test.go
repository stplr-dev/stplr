// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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

package dl_test

import (
	"context"
	"encoding/hex"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/pkg/dl"
)

func TestGitDownloaderMatchUrl(t *testing.T) {
	d := dl.GitDownloader{}
	assert.True(t, d.MatchURL("git+https://example.com/org/project.git"))
	assert.False(t, d.MatchURL("https://example.com/org/project.git"))
}

func TestGitDownloaderDownload(t *testing.T) {
	d := dl.GitDownloader{}

	createTempDir := func(t *testing.T, name string) string {
		t.Helper()
		dir, err := os.MkdirTemp("", "test-"+name)
		assert.NoError(t, err)
		t.Cleanup(func() {
			_ = os.RemoveAll(dir)
		})
		return dir
	}

	t.Run("simple", func(t *testing.T) {
		dest := createTempDir(t, "simple")

		dlType, name, err := d.Download(context.Background(), dl.Options{
			URL:         "git+https://altlinux.space/stapler/repo-for-tests.git",
			Destination: dest,
		})

		assert.NoError(t, err)
		assert.Equal(t, dl.TypeDir, dlType)
		assert.Equal(t, "repo-for-tests", name)
	})

	t.Run("with hash", func(t *testing.T) {
		dest := createTempDir(t, "with-hash")

		hsh, err := hex.DecodeString("33c912b855352663550003ca6b948ae3df1f38e2c036f5a85775df5967e143bf")
		assert.NoError(t, err)

		dlType, name, err := d.Download(context.Background(), dl.Options{
			URL:           "git+https://altlinux.space/stapler/repo-for-tests.git?~rev=init&~name=test",
			Destination:   dest,
			Hash:          hsh,
			HashAlgorithm: "sha256",
		})

		assert.NoError(t, err)
		assert.Equal(t, dl.TypeDir, dlType)
		assert.Equal(t, "test", name)
	})

	t.Run("with hash (checksum mismatch)", func(t *testing.T) {
		dest := createTempDir(t, "with-hash-checksum-mismatch")

		hsh, err := hex.DecodeString("33c912b855352663550003ca6b948ae3df1f38e2c036f5a85775df5967e143bf")
		assert.NoError(t, err)

		_, _, err = d.Download(context.Background(), dl.Options{
			URL:           "git+https://altlinux.space/stapler/repo-for-tests.git",
			Destination:   dest,
			Hash:          hsh,
			HashAlgorithm: "sha256",
		})

		assert.ErrorIs(t, err, dl.ErrChecksumMismatch)
	})
}

func TestGitDownloaderUpdate(t *testing.T) {
	d := dl.GitDownloader{}

	createTempDir := func(t *testing.T, name string) string {
		t.Helper()
		dir, err := os.MkdirTemp("", "test-"+name)

		assert.NoError(t, err)
		t.Cleanup(func() {
			_ = os.RemoveAll(dir)
		})
		return dir
	}

	setupOldRepo := func(t *testing.T, dest string) {
		t.Helper()

		cmd := exec.Command("git", "clone", "https://altlinux.space/stapler/repo-for-tests.git", dest)
		err := cmd.Run()
		assert.NoError(t, err)

		cmd = exec.Command("git", "-C", dest, "reset", "--hard", "init")
		err = cmd.Run()
		assert.NoError(t, err)
	}

	t.Run("simple", func(t *testing.T) {
		dest := createTempDir(t, "update")

		setupOldRepo(t, dest)

		cmd := exec.Command("git", "-C", dest, "rev-parse", "HEAD")
		oldHash, err := cmd.Output()
		assert.NoError(t, err)

		updated, err := d.Update(dl.Options{
			URL:         "git+https://altlinux.space/stapler/repo-for-tests.git",
			Destination: dest,
		})

		assert.NoError(t, err)
		assert.True(t, updated)

		cmd = exec.Command("git", "-C", dest, "rev-parse", "HEAD")
		newHash, err := cmd.Output()
		assert.NoError(t, err)
		assert.NotEqual(t, string(oldHash), string(newHash), "Repository should be updated")
	})

	t.Run("with hash", func(t *testing.T) {
		dest := createTempDir(t, "update")

		setupOldRepo(t, dest)

		hsh, err := hex.DecodeString("0dc4f3c68c435d0cd7a5ee960f965815fa9c4ee0571839cdb8f9de56e06f91eb")
		assert.NoError(t, err)

		updated, err := d.Update(dl.Options{
			URL:           "git+https://altlinux.space/stapler/repo-for-tests.git?~rev=test-update-git-downloader",
			Destination:   dest,
			Hash:          hsh,
			HashAlgorithm: "sha256",
		})

		assert.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("with hash (checksum mismatch)", func(t *testing.T) {
		dest := createTempDir(t, "update")

		setupOldRepo(t, dest)

		hsh, err := hex.DecodeString("33c912b855352663550003ca6b948ae3df1f38e2c036f5a85775df5967e143bf")
		assert.NoError(t, err)

		_, err = d.Update(dl.Options{
			URL:           "git+https://altlinux.space/stapler/repo-for-tests.git?~rev=test-update-git-downloader",
			Destination:   dest,
			Hash:          hsh,
			HashAlgorithm: "sha256",
		})

		assert.ErrorIs(t, err, dl.ErrChecksumMismatch)
	})
}
