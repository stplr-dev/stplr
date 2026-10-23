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

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"io"
	"os"
	"os/exec"
	"strings"
)

// exportDataLookup resolves the input file's imports to compiler export data
// via `go list -export`. Unlike the "source" importer, this reuses the build
// cache instead of type-checking every transitive dependency from scratch.
func exportDataLookup(file *ast.File) (importer.Lookup, error) {
	if len(file.Imports) == 0 {
		return nil, nil
	}

	args := []string{"list", "-export", "-deps", "-f", "{{.ImportPath}}\t{{.Export}}"}
	for _, imp := range file.Imports {
		args = append(args, strings.Trim(imp.Path.Value, `"`))
	}

	var stderr bytes.Buffer
	cmd := exec.Command("go", args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list -export: %w\n%s", err, stderr.String())
	}

	exports := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		path, export, ok := strings.Cut(sc.Text(), "\t")
		if ok && export != "" {
			exports[path] = export
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	return func(path string) (io.ReadCloser, error) {
		export, ok := exports[path]
		if !ok {
			return nil, fmt.Errorf("no export data for %q", path)
		}
		return os.Open(export)
	}, nil
}
