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
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MethodInfo struct {
	Name       string
	Params     []ParamInfo
	Results    []ResultInfo
	EntityName string
}

type ParamInfo struct {
	Name string
	Type string
}

type ResultInfo struct {
	Name  string
	Type  string
	Index int
}

func extractImports(node *ast.File) []string {
	var imports []string
	for _, imp := range node.Imports {
		if imp.Path.Value != "" {
			imports = append(imports, imp.Path.Value)
		}
	}
	return imports
}

func output(path string, buf bytes.Buffer) {
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

func main() {
	path := os.Getenv("GOFILE")
	if path == "" {
		log.Fatal("GOFILE must be set")
	}

	if len(os.Args) < 2 {
		log.Fatal("At least one entity name must be provided")
	}

	entityNames := os.Args[1:]

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		log.Fatalf("parsing file: %v", err)
	}

	packageName := node.Name.Name

	// Find all specified entities
	entityData := make(map[string][]*ast.Field)

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			for _, entityName := range entityNames {
				if typeSpec.Name.Name == entityName {
					interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok {
						log.Fatalf("entity %s is not an interface", entityName)
					}
					entityData[entityName] = interfaceType.Methods.List
				}
			}
		}
	}

	// Verify all entities were found
	for _, entityName := range entityNames {
		if _, found := entityData[entityName]; !found {
			log.Fatalf("interface %s not found", entityName)
		}
	}

	var buf bytes.Buffer

	buf.WriteString(`
// DO NOT EDIT MANUALLY. This file is generated.

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


`)

	buf.WriteString(fmt.Sprintf("package %s\n", packageName))

	// Generate base structures for all entities
	baseStructs(&buf, entityNames, extractImports(node))

	// Generate method-specific code for each entity
	for _, entityName := range entityNames {
		methods := parseMethodsFromFields(entityName, entityData[entityName])
		argsGen(&buf, methods)
	}

	output(path, buf)
}

func parseMethodsFromFields(entityName string, fields []*ast.Field) []MethodInfo {
	var methods []MethodInfo

	for _, field := range fields {
		if len(field.Names) == 0 {
			continue
		}

		methodName := field.Names[0].Name
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		method := MethodInfo{
			Name:       methodName,
			EntityName: entityName,
		}

		// Parse parameters, excluding context.Context
		if funcType.Params != nil {
			for i, param := range funcType.Params.List {
				paramType := typeToString(param.Type)
				// Skip context.Context parameters
				if paramType == "context.Context" {
					continue
				}
				if len(param.Names) == 0 {
					method.Params = append(method.Params, ParamInfo{
						Name: fmt.Sprintf("Arg%d", i),
						Type: paramType,
					})
				} else {
					for _, name := range param.Names {
						method.Params = append(method.Params, ParamInfo{
							Name: cases.Title(language.Und, cases.NoLower).String(name.Name),
							Type: paramType,
						})
					}
				}
			}
		}

		// Parse results
		if funcType.Results != nil {
			resultIndex := 0
			for _, result := range funcType.Results.List {
				resultType := typeToString(result.Type)
				if resultType == "error" {
					continue // Skip error in response struct
				}

				if len(result.Names) == 0 {
					method.Results = append(method.Results, ResultInfo{
						Name:  fmt.Sprintf("Result%d", resultIndex),
						Type:  resultType,
						Index: resultIndex,
					})
				} else {
					for _, name := range result.Names {
						method.Results = append(method.Results, ResultInfo{
							Name:  cases.Title(language.Und, cases.NoLower).String(name.Name),
							Type:  resultType,
							Index: resultIndex,
						})
					}
				}
				resultIndex++
			}
		}

		methods = append(methods, method)
	}

	return methods
}

func argsGen(buf *bytes.Buffer, methods []MethodInfo) {
	// Add template functions first
	funcMap := template.FuncMap{
		"lowerFirst": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
		"zeroValue": func(typeName string) string {
			typeName = strings.TrimSpace(typeName)

			switch typeName {
			case "string":
				return "\"\""
			case "int", "int8", "int16", "int32", "int64":
				return "0"
			case "uint", "uint8", "uint16", "uint32", "uint64":
				return "0"
			case "float32", "float64":
				return "0.0"
			case "bool":
				return "false"
			}

			if strings.HasPrefix(typeName, "*") {
				return "nil"
			}
			if strings.HasPrefix(typeName, "[]") ||
				strings.HasPrefix(typeName, "map[") ||
				strings.HasPrefix(typeName, "chan ") {
				return "nil"
			}

			if typeName == "interface{}" {
				return "nil"
			}

			// If external type: pkg.Type
			if strings.Contains(typeName, ".") {
				return typeName + "{}"
			}

			// If starts with uppercase — likely struct
			if len(typeName) > 0 && unicode.IsUpper(rune(typeName[0])) {
				return typeName + "{}"
			}

			return "nil"
		},
	}

	argsTemplate := template.Must(template.New("args").Funcs(funcMap).Parse(`
{{range .}}
type {{.EntityName}}{{.Name}}Args struct {
{{range .Params}}	{{.Name}} {{.Type}}
{{end}}}

type {{.EntityName}}{{.Name}}Resp struct {
{{range .Results}}	{{.Name}} {{.Type}}
{{end}}}

func (s *{{.EntityName}}RPC) {{.Name}}(ctx context.Context, {{range $i, $p := .Params}}{{if $i}}, {{end}}{{lowerFirst $p.Name}} {{$p.Type}}{{end}}) ({{range $i, $r := .Results}}{{if $i}}, {{end}}{{$r.Type}}{{end}}{{if .Results}}, {{end}}error) {
	var resp *{{.EntityName}}{{.Name}}Resp
	err := s.client.Call(ctx, "Plugin.{{.Name}}", &{{.EntityName}}{{.Name}}Args{
{{range .Params}}		{{.Name}}: {{lowerFirst .Name}},
{{end}}	}, &resp)
	if err != nil {
		return {{range $i, $r := .Results}}{{if $i}}, {{end}}{{zeroValue $r.Type}}{{end}}{{if .Results}}, {{end}}err
	}
	return {{range $i, $r := .Results}}{{if $i}}, {{end}}resp.{{$r.Name}}{{end}}{{if .Results}}, {{end}}nil
}

func (s *{{.EntityName}}RPCServer) {{.Name}}(ctx context.Context, args *{{.EntityName}}{{.Name}}Args, resp *{{.EntityName}}{{.Name}}Resp) error {
	{{if .Results}}{{range $i, $r := .Results}}{{if $i}}, {{end}}{{lowerFirst $r.Name}}{{end}}, err := {{else}}err := {{end}}s.Impl.{{.Name}}(ctx,{{range $i, $p := .Params}}{{if $i}}, {{end}}args.{{$p.Name}}{{end}})
	if err != nil {
		return err
	}
	{{if .Results}}*resp = {{.EntityName}}{{.Name}}Resp{
{{range .Results}}		{{.Name}}: {{lowerFirst .Name}},
{{end}}	}
	{{else}}*resp = {{.EntityName}}{{.Name}}Resp{}
	{{end}}return nil
}
{{end}}
`))

	err := argsTemplate.Execute(buf, methods)
	if err != nil {
		log.Fatalf("execute args template: %v", err)
	}
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		xStr := typeToString(t.X)
		if xStr == "context" && t.Sel.Name == "Context" {
			return "context.Context"
		}
		return xStr + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "interface{}"
	}
}

func baseStructs(buf *bytes.Buffer, entityNames, imports []string) {
	// Ensure "context" is included in imports
	updatedImports := imports
	hasContext := false
	for _, imp := range imports {
		if strings.Contains(imp, `"context"`) {
			hasContext = true
			break
		}
	}
	if !hasContext {
		updatedImports = append(updatedImports, `"context"`)
	}

	contentTemplate := template.Must(template.New("").Parse(`
import (
	"github.com/keegancsmith/rpc"

	"github.com/hashicorp/go-plugin"
{{range .Imports}}	{{.}}
{{end}}
)

{{range .EntityNames}}
type {{ . }}Plugin struct {
	Impl {{ . }}
}

type {{ . }}RPCServer struct {
	Impl {{ . }}
}

type {{ . }}RPC struct {
	client *rpc.Client
}

func (p *{{ . }}Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &{{ . }}RPC{client: c}, nil
}

func (p *{{ . }}Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &{{ . }}RPCServer{Impl: p.Impl}, nil
}

{{end}}
`))
	err := contentTemplate.Execute(buf, struct {
		EntityNames []string
		Imports     []string
	}{
		EntityNames: entityNames,
		Imports:     updatedImports,
	})
	if err != nil {
		log.Fatalf("execute template: %v", err)
	}
}
