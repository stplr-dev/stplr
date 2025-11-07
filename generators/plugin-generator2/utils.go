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

package main

import (
	"go/ast"
	"strings"
	"unicode"

	"github.com/dave/jennifer/jen"
)

func extractImportsWithAlias(node *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		var name string
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}
		imports[name] = path
	}
	return imports
}

func ensureImport(imports map[string]string, name, path string) {
	// Если импорт уже есть — ничего не делаем
	if _, exists := imports[name]; exists {
		return
	}
	imports[name] = path
}

func jenType(t string, imports map[string]string) *jen.Statement {
	switch {
	case strings.HasPrefix(t, "*"):
		return jen.Op("*").Add(jenType(t[1:], imports))
	case strings.HasPrefix(t, "[]"):
		return jen.Index().Add(jenType(t[2:], imports))
	case strings.Contains(t, "."):
		parts := strings.Split(t, ".")
		if len(parts) == 2 {
			return jen.Qual(imports[parts[0]], parts[1])
		}
	}
	return jen.Id(t)
}

func jenAddZeroValue(t string) jen.Code {
	t = strings.TrimSpace(t)

	switch t {
	case "string":
		return jen.Lit("")
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return jen.Lit(0)
	case "bool":
		return jen.False()
	case "interface{}":
		return jen.Nil()
	}

	if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") ||
		strings.HasPrefix(t, "map[") || strings.HasPrefix(t, "chan ") {
		return jen.Nil()
	}

	if strings.Contains(t, ".") {
		return jen.Op(t + "{}")
	}

	if len(t) > 0 && unicode.IsUpper(rune(t[0])) {
		return jen.Op(t + "{}")
	}

	return jen.Nil()
}
