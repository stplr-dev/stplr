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
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/cliutils"
)

type useCase struct {
	out output.Output
}

func New(out output.Output) *useCase {
	return &useCase{
		out,
	}
}

func (u *useCase) Run(ctx context.Context) error {
	archivePath := "stplr-support.tar.gz"

	if err := createSupportArchive(ctx, archivePath); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Failed to generate support archive"))
	}

	u.out.Info("Support archive %s has been created. You can send it to whoever needs it.", archivePath)

	return nil
}

func createSupportArchive(ctx context.Context, archivePath string) error {
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := copyFileToTar(tw, "/etc/os-release"); err != nil {
		return fmt.Errorf("add os-release: %w", err)
	}

	if err := addFilteredConfig(tw, "/etc/stplr/stplr.toml"); err != nil {
		return fmt.Errorf("add filtered stplr.toml: %w", err)
	}

	if err := executeCommands(ctx, tw); err != nil {
		return fmt.Errorf("execute commands: %w", err)
	}

	if err := addDiskUsage(tw); err != nil {
		return fmt.Errorf("add disk usage: %w", err)
	}

	if err := addStplrLogs(tw); err != nil {
		return fmt.Errorf("add stplr logs: %w", err)
	}

	return nil
}

func executeCommands(ctx context.Context, tw *tar.Writer) error {
	cmds := [][]string{
		{"ls", "-la", "/var/cache/stplr"},
		{"ls", "-la", "/var/cache/stplr/repo"},
		{"ls", "-la", "/etc/stplr"},
		{"cat", "/proc/sys/kernel/unprivileged_userns_clone"},
	}

	var buf bytes.Buffer
	for _, cmdArgs := range cmds {
		buf.WriteString(fmt.Sprintf(">>> %s\n", strings.Join(cmdArgs, " ")))
		//gosec:disable G204 -- Expected
		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
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

func addStplrLogs(tw *tar.Writer) error {
	executable, err := os.Executable()
	if err != nil {
		return cliutils.FormatCliExit("failed to get executable path", err)
	}

	r, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Since: time.Duration(-72) * time.Hour,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_EXE,
				Value: executable,
			},
		},
	})
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

func addFilteredConfig(tw *tar.Writer, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "url") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				lines[i] = parts[0] + ` = "<filtered>"`
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

func copyFileToTar(tw *tar.Writer, path string) error {
	info, err := os.Stat(path)
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

	f, err := os.Open(path)
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

func addDiskUsage(tw *tar.Writer) error {
	path := "/var/cache/stplr"

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("statfs failed: %w", err)
	}

	if stat.Bsize < 0 {
		return fmt.Errorf("invalid block size: %d", stat.Bsize)
	}
	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	free := stat.Bavail * blockSize
	used := total - free

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Disk usage for %s:\n", path)
	fmt.Fprintf(buf, "Total: %d bytes\n", total)
	fmt.Fprintf(buf, "Used:  %d bytes\n", used)
	fmt.Fprintf(buf, "Free:  %d bytes\n", free)

	return addReaderToTar(tw, "disk-usage.log", bytes.NewReader(buf.Bytes()), int64(buf.Len()))
}
