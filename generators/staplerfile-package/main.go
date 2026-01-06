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
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

func resolvedStructGenerator(buf *bytes.Buffer, fields []*ast.Field) {
	contentTemplate := template.Must(template.New("").Parse(`
type {{ .EntityNameLower }}Resolved struct {
{{ .StructFields }}
}
`))

	var structFieldsBuilder strings.Builder

	for _, field := range fields {
		for _, name := range field.Names {
			fieldTypeStr := exprToString(field.Type)

			var buf bytes.Buffer
			buf.WriteString("\t")
			buf.WriteString(name.Name)
			buf.WriteString(" ")
			buf.WriteString(fieldTypeStr)

			jsonTag := ""
			if field.Tag != nil {
				raw := strings.Trim(field.Tag.Value, "`")
				tag := reflect.StructTag(raw)
				if val := tag.Get("json"); val != "" {
					jsonTag = val
				}
			}
			if jsonTag == "" {
				jsonTag = strings.ToLower(name.Name)
			}
			buf.WriteString(fmt.Sprintf(" `json:\"%s\"`", jsonTag))
			buf.WriteString("\n")
			structFieldsBuilder.Write(buf.Bytes())
		}
	}

	params := struct {
		EntityNameLower string
		StructFields    string
	}{
		EntityNameLower: "package",
		StructFields:    structFieldsBuilder.String(),
	}

	err := contentTemplate.Execute(buf, params)
	if err != nil {
		log.Fatalf("execute template: %v", err)
	}
}

func toResolvedFuncGenerator(buf *bytes.Buffer, fields []*ast.Field) {
	contentTemplate := template.Must(template.New("").Parse(`
func {{ .EntityName }}ToResolved(src *{{ .EntityName }}) {{ .EntityNameLower }}Resolved {
	return {{ .EntityNameLower }}Resolved{
{{ .Assignments }}
	}
}
`))

	var assignmentsBuilder strings.Builder

	for _, field := range fields {
		for _, name := range field.Names {
			var assignBuf bytes.Buffer
			assignBuf.WriteString("\t\t")
			assignBuf.WriteString(name.Name)
			assignBuf.WriteString(": ")
			if isOverridableField(field.Type) {
				assignBuf.WriteString(fmt.Sprintf("src.%s.Resolved()", name.Name))
			} else {
				assignBuf.WriteString(fmt.Sprintf("src.%s", name.Name))
			}
			assignBuf.WriteString(",\n")
			assignmentsBuilder.Write(assignBuf.Bytes())
		}
	}

	params := struct {
		EntityName      string
		EntityNameLower string
		Assignments     string
	}{
		EntityName:      "Package",
		EntityNameLower: "package",
		Assignments:     assignmentsBuilder.String(),
	}

	err := contentTemplate.Execute(buf, params)
	if err != nil {
		log.Fatalf("execute template: %v", err)
	}
}

func resolveFuncGenerator(buf *bytes.Buffer, fields []*ast.Field) {
	contentTemplate := template.Must(template.New("").Parse(`
func Resolve{{ .EntityName }}(pkg *{{ .EntityName }}, overrides []string) {
{{.Code}}}
`))

	var codeBuilder strings.Builder

	for _, field := range fields {
		for _, name := range field.Names {
			if isOverridableField(field.Type) {
				var buf bytes.Buffer
				buf.WriteString(fmt.Sprintf("\t\tpkg.%s.Resolve(overrides)\n", name.Name))
				codeBuilder.Write(buf.Bytes())
			}
		}
	}

	params := struct {
		EntityName string
		Code       string
	}{
		EntityName: "Package",
		Code:       codeBuilder.String(),
	}

	err := contentTemplate.Execute(buf, params)
	if err != nil {
		log.Fatalf("execute template: %v", err)
	}
}

func celColumnMapGenerator(buf *bytes.Buffer, fields []*ast.Field) {
	contentTemplate := template.Must(template.New("").Parse(`
// GetCELColumnMap returns a map of CEL field names to their SQL column information
func GetCELColumnMap() map[string]cel2sqlite.ColumnInfo {
	return map[string]cel2sqlite.ColumnInfo{
{{ .Entries }}
	}
}
`))

	var entriesBuilder strings.Builder

	for _, field := range fields {
		for _, name := range field.Names {
			// Получаем теги
			var celName, sqlName string
			var skip bool

			if field.Tag != nil {
				raw := strings.Trim(field.Tag.Value, "`")
				tag := reflect.StructTag(raw)

				// Получаем CEL имя из тега sh
				celName = tag.Get("sh")
				if celName == "" {
					celName = strings.ToLower(name.Name)
				}

				// Получаем SQL имя из тега xorm
				xormTag := tag.Get("xorm")
				sqlName = extractSQLName(xormTag)

				// Пропускаем поля с xorm:"-"
				if xormTag == "-" || sqlName == "" {
					skip = true
				}
			} else {
				celName = strings.ToLower(name.Name)
				skip = true
			}

			if skip {
				continue
			}

			// Определяем тип колонки
			colType := determineColumnType(field.Type)

			var entryBuf bytes.Buffer
			entryBuf.WriteString(fmt.Sprintf("\t\t\"%s\": {SQLName: \"%s\", Type: cel2sqlite.%s},\n",
				celName, sqlName, colType))
			entriesBuilder.Write(entryBuf.Bytes())
		}
	}

	params := struct {
		Entries string
	}{
		Entries: entriesBuilder.String(),
	}

	err := contentTemplate.Execute(buf, params)
	if err != nil {
		log.Fatalf("execute template: %v", err)
	}
}

func extractSQLName(xormTag string) string {
	if xormTag == "" || xormTag == "-" {
		return ""
	}

	parts := strings.Split(xormTag, " ")
	for _, part := range parts {
		part = strings.Trim(part, "'\"")
		// Пропускаем ключевые слова xorm
		if part == "pk" || part == "notnull" || part == "json" ||
			strings.HasPrefix(part, "default") || strings.Contains(part, ":") {
			continue
		}
		// Первый не-ключевой элемент это имя колонки
		if part != "" {
			return part
		}
	}
	return ""
}

func determineColumnType(expr ast.Expr) string {
	switch overridableFieldKind(expr) {
	case "overridable_array":
		return "ColumnTypeOverridableFieldArray"
	case "overridable_scalar":
		return "ColumnTypeOverridableField"
	}

	typeStr := exprToString(expr)

	switch {
	case typeStr == "string":
		return "ColumnTypeString"
	case typeStr == "int" || typeStr == "uint" ||
		typeStr == "int64" || typeStr == "uint64" ||
		typeStr == "int32" || typeStr == "uint32":
		return "ColumnTypeInt"
	case typeStr == "bool":
		return "ColumnTypeBool"
	case strings.HasPrefix(typeStr, "[]"):
		return "ColumnTypeJSONArray"
	default:
		return "ColumnTypeString"
	}
}

func main() {
	path := os.Getenv("GOFILE")
	if path == "" {
		log.Fatal("GOFILE must be set")
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		log.Fatalf("parsing file: %v", err)
	}

	entityName := "Package"

	found := false
	fields := make([]*ast.Field, 0)

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			if typeSpec.Name.Name != entityName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			fields = structType.Fields.List
			found = true
		}
	}

	if !found {
		log.Fatalf("struct %s not found", entityName)
	}

	var buf bytes.Buffer

	buf.WriteString("// DO NOT EDIT MANUALLY. This file is generated.\n")
	buf.WriteString("package staplerfile\n\n")
	buf.WriteString("import \"go.stplr.dev/stplr/internal/cel2sqlite\"\n")

	resolvedStructGenerator(&buf, fields)
	toResolvedFuncGenerator(&buf, fields)
	resolveFuncGenerator(&buf, fields)
	celColumnMapGenerator(&buf, fields)

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("formatting: %v", err)
	}

	outPath := strings.TrimSuffix(path, ".go") + "_gen.go"
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create file: %v", err)
	}
	_, err = outFile.Write(formatted)
	if err != nil {
		log.Fatalf("writing output: %v", err)
	}
	outFile.Close()
}

func exprToString(expr ast.Expr) string {
	if t, ok := expr.(*ast.IndexExpr); ok {
		if ident, ok := t.X.(*ast.Ident); ok && ident.Name == "OverridableField" {
			return exprToString(t.Index)
		}
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return "<invalid>"
	}
	return buf.String()
}

func overridableFieldKind(expr ast.Expr) string {
	indexExpr, ok := expr.(*ast.IndexExpr)
	if !ok {
		return ""
	}

	ident, ok := indexExpr.X.(*ast.Ident)
	if !ok || ident.Name != "OverridableField" {
		return ""
	}

	switch indexExpr.Index.(type) {
	case *ast.ArrayType:
		return "overridable_array"
	default:
		return "overridable_scalar"
	}
}

func isOverridableField(expr ast.Expr) bool {
	return overridableFieldKind(expr) != ""
}
