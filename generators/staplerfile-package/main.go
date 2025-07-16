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
			// Поле с типом
			fieldTypeStr := exprToString(field.Type)

			// Структура поля
			var buf bytes.Buffer
			buf.WriteString("\t")
			buf.WriteString(name.Name)
			buf.WriteString(" ")
			buf.WriteString(fieldTypeStr)

			// Обработка json-тега
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

	entityName := "Package" // имя структуры, которую анализируем

	found := false

	fields := make([]*ast.Field, 0)

	// Ищем структуру с нужным именем
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
	buf.WriteString("package staplerfile")

	resolvedStructGenerator(&buf, fields)
	toResolvedFuncGenerator(&buf, fields)
	resolveFuncGenerator(&buf, fields)

	// Форматируем вывод
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
			return exprToString(t.Index) // T
		}
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return "<invalid>"
	}
	return buf.String()
}

func isOverridableField(expr ast.Expr) bool {
	indexExpr, ok := expr.(*ast.IndexExpr)
	if !ok {
		return false
	}
	ident, ok := indexExpr.X.(*ast.Ident)
	return ok && ident.Name == "OverridableField"
}
