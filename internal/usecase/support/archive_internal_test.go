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

package support

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestAddFilteredConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fs := afero.NewMemMapFs()
	configContent := `[[repo]]
name = 'stplr-repo'
ref = ''
url = 'https://example.com/secret-url-with-token.git'`

	require.NoError(t, afero.WriteFile(fs, "/etc/stplr/stplr.toml", []byte(configContent), 0o644))

	creator := archiveCreator{
		fs,
		NewMockCommandExecutor(ctrl),
		NewMockJournalReader(ctrl),
		NewMockDiskStatter(ctrl),
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := creator.addFilteredConfig(tw, "/etc/stplr/stplr.toml")
	require.NoError(t, err)

	require.NoError(t, tw.Close())

	tr := tar.NewReader(&buf)
	header, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "stplr.toml.filtered", header.Name)

	content, err := io.ReadAll(tr)
	require.NoError(t, err)

	assert.Contains(t, string(content), `[[repo]]
name = 'stplr-repo'
ref = ''
url = "<filtered>"`)
}

func TestCopyFileToTar(t *testing.T) {
	fs := afero.NewMemMapFs()
	testContent := []byte("test file content")
	afero.WriteFile(fs, "/etc/os-release", testContent, 0o644)

	creator := &archiveCreator{
		fs: fs,
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := creator.copyFileToTar(tw, "/etc/os-release")
	require.NoError(t, err)

	tw.Close()

	tr := tar.NewReader(&buf)
	header, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "os-release", header.Name)

	content, err := io.ReadAll(tr)
	require.NoError(t, err)
	assert.Equal(t, testContent, content)
}

func TestAddDiskUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDisk := NewMockDiskStatter(ctrl)
	mockDisk.EXPECT().Statfs("/var/cache/stplr").
		Return(uint64(10000000), uint64(6000000), uint64(4000000), nil)

	creator := &archiveCreator{
		disk: mockDisk,
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := creator.addDiskUsage(tw)
	require.NoError(t, err)

	tw.Close()

	tr := tar.NewReader(&buf)
	header, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "disk-usage.log", header.Name)

	content, err := io.ReadAll(tr)
	require.NoError(t, err)

	assert.Contains(t, string(content), "Total: 10000000 bytes")
	assert.Contains(t, string(content), "Used:  6000000 bytes")
	assert.Contains(t, string(content), "Free:  4000000 bytes")
}

func TestCreateSupportArchiveExtractAndVerify(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fs := afero.NewMemMapFs()

	afero.WriteFile(fs, "/etc/os-release", []byte("NAME=TestOS"), 0o644)
	afero.WriteFile(fs, "/etc/stplr/stplr.toml", []byte("url = \"secret\"\nname = \"test\""), 0o644)

	mockExecutor := NewMockCommandExecutor(ctrl)
	mockJournal := NewMockJournalReader(ctrl)
	mockDisk := NewMockDiskStatter(ctrl)

	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]byte("ok"), nil)

	mockJournal.EXPECT().Read(gomock.Any()).
		Return(io.NopCloser(bytes.NewReader([]byte("logs"))), nil)

	mockDisk.EXPECT().Statfs(gomock.Any()).
		Return(uint64(1000), uint64(500), uint64(500), nil)

	creator := &archiveCreator{
		fs:       fs,
		executor: mockExecutor,
		journal:  mockJournal,
		disk:     mockDisk,
	}

	archivePath := "/tmp/test.tar.gz"
	err := creator.CreateSupportArchive(context.Background(), archivePath)
	require.NoError(t, err)

	archiveFile, err := fs.Open(archivePath)
	require.NoError(t, err)
	defer archiveFile.Close()

	gzr, err := gzip.NewReader(archiveFile)
	require.NoError(t, err)
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	files := make(map[string]string)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		content, err := io.ReadAll(tr)
		require.NoError(t, err)

		files[header.Name] = string(content)
	}

	assert.Contains(t, files, "os-release")
	assert.Contains(t, files, "stplr.toml.filtered")
	assert.Contains(t, files, "commands.log")
	assert.Contains(t, files, "journal.log")
	assert.Contains(t, files, "disk-usage.log")

	assert.Contains(t, files["stplr.toml.filtered"], `url = "<filtered>"`)
	assert.NotContains(t, files["stplr.toml.filtered"], "secret")
}
