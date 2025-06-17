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
	"os"
	"path/filepath"
	"testing"

	"github.com/goreleaser/nfpm/v2/files"
	"github.com/stretchr/testify/assert"

	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/alrsh"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/types"
)

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		expected    bool
	}{
		{"usr/bin binary", "/usr/bin/test", true},
		{"bin binary", "/bin/test", true},
		{"usr/local/bin binary", "/usr/local/bin/test", true},
		{"lib file", "/usr/lib/test.so", false},
		{"etc file", "/etc/config", false},
		{"empty destination", "", false},
		{"root level file", "./test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryFile(tt.destination)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSafeName(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		expected    string
		expectError bool
	}{
		{"usr/bin path", "./usr/bin/test", "_usr_bin_test", false},
		{"bin path", "./bin/test", "_bin_test", false},
		{"nested path", "./usr/local/bin/app", "_usr_local_bin_app", false},
		{"path with spaces", "./usr/bin/my app", "_usr_bin_my app", false},
		{"empty after trim", ".", "", true},
		{"empty string", "", "", true},
		{"only dots", "..", ".", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateSafeName(tt.destination)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCreateWrapperScript(t *testing.T) {
	tests := []struct {
		name            string
		origFilePath    string
		profilePath     string
		expectedContent string
	}{
		{
			"basic wrapper",
			"/usr/lib/alr/firejailed/_usr_bin_test",
			"/usr/lib/alr/firejailed/_usr_bin_test.profile",
			"#!/bin/bash\nexec firejail --profile=\"/usr/lib/alr/firejailed/_usr_bin_test.profile\" \"/usr/lib/alr/firejailed/_usr_bin_test\" \"$@\"\n",
		},
		{
			"path with spaces",
			"/usr/lib/alr/firejailed/_usr_bin_my_app",
			"/usr/lib/alr/firejailed/_usr_bin_my_app.profile",
			"#!/bin/bash\nexec firejail --profile=\"/usr/lib/alr/firejailed/_usr_bin_my_app.profile\" \"/usr/lib/alr/firejailed/_usr_bin_my_app\" \"$@\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scriptPath := filepath.Join(tmpDir, "wrapper.sh")

			err := createWrapperScript(scriptPath, tt.origFilePath, tt.profilePath)

			assert.NoError(t, err)
			assert.FileExists(t, scriptPath)

			content, err := os.ReadFile(scriptPath)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, string(content))

			// Check file permissions
			info, err := os.Stat(scriptPath)
			assert.NoError(t, err)
			assert.Equal(t, os.FileMode(defaultDirMode), info.Mode())
		})
	}
}

func TestCreateFirejailedBinary(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) (*alrsh.Package, *files.Content, types.Directories)
		expectError bool
		errorMsg    string
	}{
		{
			"successful creation with default profile",
			func(tmpDir string) (*alrsh.Package, *files.Content, types.Directories) {
				pkgDir := filepath.Join(tmpDir, "pkg")
				scriptDir := filepath.Join(tmpDir, "scripts")
				os.MkdirAll(pkgDir, 0o755)
				os.MkdirAll(scriptDir, 0o755)

				binDir := filepath.Join(pkgDir, "usr", "bin")
				os.MkdirAll(binDir, 0o755)

				srcBinary := filepath.Join(binDir, "test-binary")
				os.WriteFile(srcBinary, []byte("#!/bin/bash\necho test"), 0o755)

				defaultProfile := filepath.Join(scriptDir, "default.profile")
				os.WriteFile(defaultProfile, []byte("include /etc/firejail/default.profile\nnet none"), 0o644)

				pkg := &alrsh.Package{
					Name: "test-pkg",
					FireJailProfiles: alrsh.OverridableFromMap(map[string]map[string]string{
						"": {"default": "default.profile"},
					}),
				}
				alrsh.ResolvePackage(pkg, []string{""})

				content := &files.Content{
					Source:      srcBinary,
					Destination: "/usr/bin/test-binary",
					Type:        "file",
				}

				dirs := types.Directories{PkgDir: pkgDir, ScriptDir: scriptDir}
				return pkg, content, dirs
			},
			false,
			"",
		},
		{
			"successful creation with specific profile",
			func(tmpDir string) (*alrsh.Package, *files.Content, types.Directories) {
				pkgDir := filepath.Join(tmpDir, "pkg")
				scriptDir := filepath.Join(tmpDir, "scripts")
				os.MkdirAll(pkgDir, 0o755)
				os.MkdirAll(scriptDir, 0o755)

				binDir := filepath.Join(pkgDir, "usr", "bin")
				os.MkdirAll(binDir, 0o755)

				srcBinary := filepath.Join(binDir, "special-binary")
				os.WriteFile(srcBinary, []byte("#!/bin/bash\necho special"), 0o755)

				defaultProfile := filepath.Join(scriptDir, "default.profile")
				os.WriteFile(defaultProfile, []byte("include /etc/firejail/default.profile"), 0o644)

				specialProfile := filepath.Join(scriptDir, "special.profile")
				os.WriteFile(specialProfile, []byte("include /etc/firejail/default.profile\nnet none\nprivate-tmp"), 0o644)

				pkg := &alrsh.Package{
					Name: "test-pkg",
					FireJailProfiles: alrsh.OverridableFromMap(map[string]map[string]string{
						"": {"default": "default.profile", "/usr/bin/special-binary": "special.profile"},
					}),
				}
				alrsh.ResolvePackage(pkg, []string{""})

				content := &files.Content{
					Source:      srcBinary,
					Destination: "/usr/bin/special-binary",
					Type:        "file",
				}

				dirs := types.Directories{PkgDir: pkgDir, ScriptDir: scriptDir}
				return pkg, content, dirs
			},
			false,
			"",
		},
		{
			"missing default profile",
			func(tmpDir string) (*alrsh.Package, *files.Content, types.Directories) {
				pkgDir := filepath.Join(tmpDir, "pkg")
				scriptDir := filepath.Join(tmpDir, "scripts")
				os.MkdirAll(pkgDir, 0o755)
				os.MkdirAll(scriptDir, 0o755)

				srcBinary := filepath.Join(tmpDir, "test-binary")
				os.WriteFile(srcBinary, []byte("#!/bin/bash\necho test"), 0o755)

				pkg := &alrsh.Package{
					Name:             "test-pkg",
					FireJailProfiles: alrsh.OverridableFromMap(map[string]map[string]string{"": {}}),
				}
				alrsh.ResolvePackage(pkg, []string{""})

				content := &files.Content{Source: srcBinary, Destination: "./usr/bin/test-binary", Type: "file"}
				dirs := types.Directories{PkgDir: pkgDir, ScriptDir: scriptDir}
				return pkg, content, dirs
			},
			true,
			"default profile is missing",
		},
		{
			"profile file not found",
			func(tmpDir string) (*alrsh.Package, *files.Content, types.Directories) {
				pkgDir := filepath.Join(tmpDir, "pkg")
				scriptDir := filepath.Join(tmpDir, "scripts")
				os.MkdirAll(pkgDir, 0o755)
				os.MkdirAll(scriptDir, 0o755)

				srcBinary := filepath.Join(tmpDir, "test-binary")
				os.WriteFile(srcBinary, []byte("#!/bin/bash\necho test"), 0o755)

				pkg := &alrsh.Package{
					Name:             "test-pkg",
					FireJailProfiles: alrsh.OverridableFromMap(map[string]map[string]string{"": {"default": "nonexistent.profile"}}),
				}
				alrsh.ResolvePackage(pkg, []string{""})

				content := &files.Content{Source: srcBinary, Destination: "./usr/bin/test-binary", Type: "file"}
				dirs := types.Directories{PkgDir: pkgDir, ScriptDir: scriptDir}
				return pkg, content, dirs
			},
			true,
			"",
		},
		{
			"invalid destination path",
			func(tmpDir string) (*alrsh.Package, *files.Content, types.Directories) {
				pkgDir := filepath.Join(tmpDir, "pkg")
				scriptDir := filepath.Join(tmpDir, "scripts")
				os.MkdirAll(pkgDir, 0o755)
				os.MkdirAll(scriptDir, 0o755)

				srcBinary := filepath.Join(tmpDir, "test-binary")
				os.WriteFile(srcBinary, []byte("#!/bin/bash\necho test"), 0o755)

				defaultProfile := filepath.Join(scriptDir, "default.profile")
				os.WriteFile(defaultProfile, []byte("include /etc/firejail/default.profile"), 0o644)

				pkg := &alrsh.Package{Name: "test-pkg", FireJailProfiles: alrsh.OverridableFromMap(map[string]map[string]string{"": {"default": "default.profile"}})}
				alrsh.ResolvePackage(pkg, []string{""})

				content := &files.Content{Source: srcBinary, Destination: ".", Type: "file"}
				dirs := types.Directories{PkgDir: pkgDir, ScriptDir: scriptDir}
				return pkg, content, dirs
			},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pkg, content, dirs := tt.setupFunc(tmpDir)

			err := createFirejailedDirectory(dirs.PkgDir)
			assert.NoError(t, err)

			result, err := createFirejailedBinary(pkg, content, dirs)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, 2)

				binContent := result[0]
				assert.Contains(t, binContent.Destination, "usr/lib/alr/firejailed/")
				assert.FileExists(t, binContent.Source)

				profileContent := result[1]
				assert.Contains(t, profileContent.Destination, "usr/lib/alr/firejailed/")
				assert.Contains(t, profileContent.Destination, ".profile")
				assert.FileExists(t, profileContent.Source)

				assert.FileExists(t, content.Source)
				wrapperBytes, err := os.ReadFile(content.Source)
				assert.NoError(t, err)
				wrapper := string(wrapperBytes)
				assert.Contains(t, wrapper, "#!/bin/bash")
				assert.Contains(t, wrapper, "firejail --profile=")
			}
		})
	}
}
