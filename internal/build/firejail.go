// ALR - Any Linux Repository
// Copyright (C) 2025 The ALR Authors
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

package build

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/goreleaser/nfpm/v2/files"
	"github.com/leonelquinteros/gotext"

	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/alrsh"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/types"
)

const (
	firejailedDir     = "/usr/lib/alr/firejailed"
	defaultDirMode    = 0o755
	defaultScriptMode = 0o755
)

var (
	ErrInvalidDestination = errors.New("invalid destination path")
	ErrMissingProfile     = errors.New("default profile is missing")
	ErrEmptyPackageName   = errors.New("package name cannot be empty")
)

var binaryDirectories = []string{
	"/usr/bin/",
	"/bin/",
	"/usr/local/bin/",
}

func moveWithSymlinkHandling(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return moveSymlink(src, dst)
	}

	if err := os.Rename(src, dst); err != nil {
		return copyAndRemove(src, dst)
	}

	return nil
}

func moveSymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("failed to read symlink: %w", err)
	}

	if err := os.Symlink(target, dst); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	if err := os.Remove(src); err != nil {
		os.Remove(dst)
		return fmt.Errorf("failed to remove original symlink: %w", err)
	}

	return nil
}

func copyAndRemove(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source stats: %w", err)
	}

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source: %w", err)
	}

	return nil
}

func moveFileWithErrorHandling(src, dst string) error {
	err := moveWithSymlinkHandling(src, dst)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: %w", err)
		}
		if os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %w", err)
		}
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil
}

func applyFirejailIntegration(
	vars *alrsh.Package,
	dirs types.Directories,
	contents []*files.Content,
) ([]*files.Content, error) {
	slog.Info(gotext.Get("Applying FireJail integration"), "package", vars.Name)

	if err := createFirejailedDirectory(dirs.PkgDir); err != nil {
		return nil, fmt.Errorf("failed to create firejailed directory: %w", err)
	}

	newContents, err := processBinaryFiles(vars, contents, dirs)
	if err != nil {
		return nil, fmt.Errorf("failed to process binary files: %w", err)
	}

	return append(contents, newContents...), nil
}

func createFirejailedDirectory(pkgDir string) error {
	firejailedPath := filepath.Join(pkgDir, firejailedDir)
	return os.MkdirAll(firejailedPath, defaultDirMode)
}

func processBinaryFiles(pkg *alrsh.Package, contents []*files.Content, dirs types.Directories) ([]*files.Content, error) {
	var newContents []*files.Content

	for _, content := range contents {
		if content.Type == "dir" {
			continue
		}

		if !isBinaryFile(content.Destination) {
			slog.Debug("content not binary file", "content", content)
			continue
		}

		slog.Debug("process content", "content", content)

		newContent, err := createFirejailedBinary(pkg, content, dirs)
		if err != nil {
			return nil, fmt.Errorf("failed to create firejailed binary for %s: %w", content.Destination, err)
		}

		if newContent != nil {
			newContents = append(newContents, newContent...)
		}
	}

	return newContents, nil
}

func isBinaryFile(destination string) bool {
	for _, binDir := range binaryDirectories {
		if strings.HasPrefix(destination, binDir) {
			return true
		}
	}
	return false
}

func createFirejailedBinary(
	pkg *alrsh.Package,
	content *files.Content,
	dirs types.Directories,
) ([]*files.Content, error) {
	origFilePath, err := generateFirejailedPath(content.Destination)
	if err != nil {
		return nil, err
	}

	profiles := pkg.FireJailProfiles.Resolved()
	sourceProfilePath, ok := profiles[content.Destination]

	if !ok {
		sourceProfilePath, ok = profiles["default"]
		if !ok {
			return nil, errors.New("default profile is missing")
		}
	}

	sourceProfilePath = filepath.Join(dirs.ScriptDir, sourceProfilePath)
	dest, err := createFirejailProfilePath(content.Destination)
	if err != nil {
		return nil, err
	}

	err = createProfile(filepath.Join(dirs.PkgDir, dest), sourceProfilePath)
	if err != nil {
		return nil, err
	}

	if err := moveFileWithErrorHandling(filepath.Join(dirs.PkgDir, content.Destination), filepath.Join(dirs.PkgDir, origFilePath)); err != nil {
		return nil, fmt.Errorf("failed to move original binary: %w", err)
	}

	content.Type = "file"
	content.Source = filepath.Join(dirs.PkgDir, content.Destination)

	// Create wrapper script
	if err := createWrapperScript(filepath.Join(dirs.PkgDir, content.Destination), origFilePath, dest); err != nil {
		return nil, fmt.Errorf("failed to create wrapper script: %w", err)
	}

	profile, err := getContentFromPath(dest, dirs.PkgDir)
	if err != nil {
		return nil, err
	}

	bin, err := getContentFromPath(origFilePath, dirs.PkgDir)
	if err != nil {
		return nil, err
	}

	return []*files.Content{
		bin,
		profile,
	}, nil
}

func getContentFromPath(path, base string) (*files.Content, error) {
	absPath := filepath.Join(base, path)

	fi, err := os.Lstat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &files.Content{
		Source:      absPath,
		Destination: path,
		FileInfo: &files.ContentFileInfo{
			MTime: fi.ModTime(),
			Mode:  fi.Mode(),
			Size:  fi.Size(),
		},
	}, nil
}

func generateSafeName(destination string) (string, error) {
	cleanPath := strings.TrimPrefix(destination, ".")
	if cleanPath == "" {
		return "", fmt.Errorf("invalid destination path: %s", destination)
	}
	return strings.ReplaceAll(cleanPath, "/", "_"), nil
}

func generateFirejailedPath(destination string) (string, error) {
	safeName, err := generateSafeName(destination)
	if err != nil {
		return "", err
	}
	return filepath.Join(firejailedDir, safeName), nil
}

func createProfile(destProfilePath, profilePath string) error {
	srcFile, err := os.Open(profilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destProfilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}

func createWrapperScript(scriptPath, origFilePath, profilePath string) error {
	scriptContent := fmt.Sprintf("#!/bin/bash\nexec firejail --profile=%q %q \"$@\"\n", profilePath, origFilePath)
	return os.WriteFile(scriptPath, []byte(scriptContent), defaultDirMode)
}

func createFirejailProfilePath(binaryPath string) (string, error) {
	name, err := generateSafeName(binaryPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(firejailedDir, fmt.Sprintf("%s.profile", name)), nil
}
