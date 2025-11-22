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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/spf13/afero"
)

type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)
}

type JournalReader interface {
	Read(since time.Duration) (io.ReadCloser, error)
}

type DiskStatter interface {
	Statfs(path string) (total, used, free uint64, err error)
}

type archiveCreator struct {
	fs       afero.Fs
	executor CommandExecutor
	journal  JournalReader
	disk     DiskStatter
}

func newArchiveCreator(fs afero.Fs) *archiveCreator {
	return &archiveCreator{
		fs:       fs,
		executor: &realCommandExecutor{},
		journal:  &realJournalReader{},
		disk:     &realDiskStatter{},
	}
}

func (ac *archiveCreator) CreateSupportArchive(ctx context.Context, archivePath string) error {
	f, err := ac.fs.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := ac.copyFileToTar(tw, "/etc/os-release"); err != nil {
		return fmt.Errorf("add os-release: %w", err)
	}

	if err := ac.addFilteredConfig(tw, "/etc/stplr/stplr.toml"); err != nil {
		return fmt.Errorf("add filtered stplr.toml: %w", err)
	}

	if err := ac.executeCommands(ctx, tw); err != nil {
		return fmt.Errorf("execute commands: %w", err)
	}

	if err := ac.addDiskUsage(tw); err != nil {
		return fmt.Errorf("add disk usage: %w", err)
	}

	if err := ac.addStplrLogs(tw); err != nil {
		return fmt.Errorf("add stplr logs: %w", err)
	}

	return nil
}

func (ac *archiveCreator) executeCommands(ctx context.Context, tw *tar.Writer) error {
	cmds := [][]string{
		{"ls", "-la", "/var/cache/stplr"},
		{"ls", "-la", "/var/cache/stplr/repo"},
		{"ls", "-la", "/etc/stplr"},
		{"cat", "/proc/sys/kernel/unprivileged_userns_clone"},
		{"uname", "-srmo"},
		{"sh", "-c", "echo XDG_CURRENT_DESKTOP=$XDG_CURRENT_DESKTOP"},
		{"sh", "-c", "echo XDG_SESSION_TYPE=$XDG_SESSION_TYPE"},
		{"sh", "-c", "echo DISPLAY=$DISPLAY"},
	}

	var buf bytes.Buffer
	for _, cmdArgs := range cmds {
		buf.WriteString(fmt.Sprintf(">>> %s\n", strings.Join(cmdArgs, " ")))
		output, err := ac.executor.Execute(ctx, cmdArgs[0], cmdArgs[1:]...)
		if err != nil {
			buf.WriteString(fmt.Sprintf("Error: %v\n", err))
		} else {
			buf.Write(output)
		}
		buf.WriteString("\n\n")
	}

	if err := addReaderToTar(tw, "commands.log", bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
		return fmt.Errorf("failed to add commands log: %w", err)
	}

	return nil
}

func (ac *archiveCreator) addStplrLogs(tw *tar.Writer) error {
	r, err := ac.journal.Read(72 * time.Hour)
	if err != nil {
		return fmt.Errorf("failed get journal: %w", err)
	}
	defer r.Close()

	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, r)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return addReaderToTar(tw, "journal.log", bytes.NewReader(buf.Bytes()), n)
}

func (ac *archiveCreator) addFilteredConfig(tw *tar.Writer, path string) error {
	data, err := afero.ReadFile(ac.fs, path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "url") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				lines[i] = parts[0] + `= "<filtered>"`
			}
		}
	}

	clean := strings.Join(lines, "\n")

	return addReaderToTar(
		tw,
		filepath.Base(path)+".filtered",
		bytes.NewReader([]byte(clean)),
		int64(len(clean)),
	)
}

func (ac *archiveCreator) copyFileToTar(tw *tar.Writer, path string) error {
	info, err := ac.fs.Stat(path)
	if err != nil {
		return err
	}

	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	hdr.Name = filepath.Base(path)

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	f, err := ac.fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

func addReaderToTar(tw *tar.Writer, name string, r io.Reader, size int64) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0o644,
		Size: size,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	_, err := io.Copy(tw, r)
	return err
}

func (ac *archiveCreator) addDiskUsage(tw *tar.Writer) error {
	path := "/var/cache/stplr"

	total, used, free, err := ac.disk.Statfs(path)
	if err != nil {
		return fmt.Errorf("statfs failed: %w", err)
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Disk usage for %s:\n", path)
	fmt.Fprintf(buf, "Total: %d bytes\n", total)
	fmt.Fprintf(buf, "Used:  %d bytes\n", used)
	fmt.Fprintf(buf, "Free:  %d bytes\n", free)

	return addReaderToTar(tw, "disk-usage.log", bytes.NewReader(buf.Bytes()), int64(buf.Len()))
}

type realCommandExecutor struct{}

func (e *realCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

type realJournalReader struct{}

func (r *realJournalReader) Read(since time.Duration) (io.ReadCloser, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	return sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Since: -since,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_EXE,
				Value: executable,
			},
		},
	})
}

type realDiskStatter struct{}

func (d *realDiskStatter) Statfs(path string) (total, used, free uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0, err
	}

	if stat.Bsize < 0 {
		return 0, 0, 0, fmt.Errorf("invalid block size: %d", stat.Bsize)
	}

	blockSize := uint64(stat.Bsize)
	total = stat.Blocks * blockSize
	free = stat.Bavail * blockSize
	used = total - free

	return total, used, free, nil
}
