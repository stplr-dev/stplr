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
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/dave/jennifer/jen"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MethodInfo struct {
	Name       string
	Params     []ParamInfo
	Results    []ResultInfo
	EntityName string
	HasContext bool
}

type ParamInfo struct {
	Name        string
	Type        string
	IsInterface bool
}

type ResultInfo struct {
	Name  string
	Type  string
	Index int
}

type EntityData struct {
	methods []string
}

type Generator struct {
	fset      *token.FileSet
	inputFile *ast.File
	pkg       *types.Package

	entityNames []string
	entities    map[string]EntityData

	f       *jen.File
	imports map[string]string
}

// prepare устанавливает значения fset, file, pkg.
func (g *Generator) prepare(input string) error {
	var err error
	g.fset = token.NewFileSet()
	g.inputFile, err = parser.ParseFile(g.fset, input, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	conf := types.Config{Importer: importer.ForCompiler(g.fset, "source", nil)}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	pkgname := g.inputFile.Name.Name

	g.pkg, err = conf.Check(pkgname, g.fset, []*ast.File{g.inputFile}, info)
	if err != nil {
		return err
	}

	g.f = jen.NewFile(pkgname)
	g.f.HeaderComment(header)

	return nil
}

func (g *Generator) prepareEntities(entityNames []string) {
	g.entityNames = entityNames
	g.entities = make(map[string]EntityData)

	for _, name := range entityNames {
		obj := g.pkg.Scope().Lookup(name)
		iface, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			log.Fatalf("entity %s is not an interface", "MyEntity")
		}

		data := EntityData{}
		data.methods = make([]string, 0)

		for i := 0; i < iface.NumMethods(); i++ {
			method := iface.Method(i)
			data.methods = append(data.methods, method.Name())
		}

		g.entities[name] = data
	}
}

const (
	keegancsmith_rpc    = "github.com/keegancsmith/rpc"
	hashicorp_go_plugin = "github.com/hashicorp/go-plugin"
)

func (g *Generator) addImports() {
	g.imports = extractImportsWithAlias(g.inputFile)
	ensureImport(g.imports, "context", "context")
	ensureImport(g.imports, "rpc", keegancsmith_rpc)
	ensureImport(g.imports, "plugin", hashicorp_go_plugin)

	for key, imp := range g.imports {
		g.f.ImportName(imp, key)
	}
}

func (g *Generator) addPluginStructs() {
	// Generate per-entity plugin structs
	for _, name := range g.entityNames {
		g.f.Type().Id(name + "Plugin").Struct(
			jen.Id("Impl").Id(name),
		)

		g.f.Type().Id(name+"RPCServer").Struct(
			jen.Id("Impl").Id(name),
			jen.Id("broker").Op("*").Qual(hashicorp_go_plugin, "MuxBroker"),
		)

		g.f.Type().Id(name+"RPC").Struct(
			jen.Id("client").Op("*").Qual(keegancsmith_rpc, "Client"),
			jen.Id("broker").Op("*").Qual(hashicorp_go_plugin, "MuxBroker"),
		)

		// Client method
		g.f.Func().Params(jen.Id("p").Op("*").Id(name+"Plugin")).Id("Client").
			Params(
				jen.Id("b").Op("*").Qual(hashicorp_go_plugin, "MuxBroker"),
				jen.Id("c").Op("*").Qual(keegancsmith_rpc, "Client"),
			).
			Params(jen.Interface(), jen.Error()).
			Block(
				jen.Return(jen.Op("&").Id(name+"RPC").Values(jen.Dict{
					jen.Id("client"): jen.Id("c"),
					jen.Id("broker"): jen.Id("b"),
				}), jen.Nil()),
			)

		// Server method
		g.f.Func().Params(jen.Id("p").Op("*").Id(name+"Plugin")).Id("Server").
			Params(jen.Id("b").Op("*").Qual(hashicorp_go_plugin, "MuxBroker")).
			Params(jen.Interface(), jen.Error()).
			Block(
				jen.Return(jen.Op("&").Id(name+"RPCServer").Values(jen.Dict{
					jen.Id("Impl"):   jen.Id("p").Dot("Impl"),
					jen.Id("broker"): jen.Id("b"),
				}), jen.Nil()),
			)
	}
}

func parseMethodsFromPackage(pkg *types.Package, entityName string) []MethodInfo {
	var methods []MethodInfo

	obj := pkg.Scope().Lookup(entityName)
	if obj == nil {
		slog.Warn("entity not found in package", "entity", entityName)
		return methods
	}

	named, ok := obj.Type().(*types.Named)
	if !ok {
		slog.Warn("entity is not a named type", "entity", entityName)
		return methods
	}

	switch t := named.Underlying().(type) {
	case *types.Interface:
		{
			for i := 0; i < t.NumMethods(); i++ {
				fn := t.Method(i)
				sig := fn.Type().(*types.Signature)
				methods = append(methods, extractMethodInfo(fn.Name(), entityName, sig))
			}
		}
	default:
		slog.Warn("unsupported underlying type", "entity", entityName, "type", fmt.Sprintf("%T", t))
	}

	return methods
}

func extractMethodInfo(name, entityName string, sig *types.Signature) MethodInfo {
	method := MethodInfo{
		Name:       name,
		EntityName: entityName,
	}

	method.Params = extractParams(sig.Params(), &method.HasContext)
	method.Results = extractResults(sig.Results())

	return method
}

func extractParams(params *types.Tuple, hasContext *bool) []ParamInfo {
	var result []ParamInfo

	for j := 0; j < params.Len(); j++ {
		v := params.At(j)
		t := v.Type()
		paramType := typeString(t)

		if paramType == "context.Context" {
			*hasContext = true
			continue
		}

		pName := normalizedName(v.Name(), "Arg", j)
		result = append(result, ParamInfo{
			Name:        pName,
			Type:        paramType,
			IsInterface: isInterfaceType(t),
		})
	}
	return result
}

func extractResults(results *types.Tuple) []ResultInfo {
	var result []ResultInfo

	for k := 0; k < results.Len(); k++ {
		v := results.At(k)
		resultType := typeString(v.Type())

		if resultType == "error" {
			continue
		}

		rName := normalizedName(v.Name(), "Result", k)
		result = append(result, ResultInfo{
			Name:  rName,
			Type:  resultType,
			Index: k,
		})
	}
	return result
}

func normalizedName(name, prefix string, index int) string {
	if name == "" {
		return fmt.Sprintf("%s%d", prefix, index)
	}
	return cases.Title(language.Und, cases.NoLower).String(name)
}

func isInterfaceType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

func typeString(t types.Type) string {
	return types.TypeString(t, func(pkg *types.Package) string {
		if pkg == nil {
			return ""
		}
		return pkg.Name()
	})
}

func (g *Generator) output(inputPath string) {
	outPath := strings.TrimSuffix(inputPath, ".go") + "_gen.go"
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create file: %v", err)
	}

	// formatted, err := format.Source(buf.Bytes())
	// if err != nil {
	// 	log.Fatalf("formatting: %v", err)
	// }

	_, err = fmt.Fprintf(outFile, "%#v", g.f)
	if err != nil {
		log.Fatalf("writing output: %v", err)
	}
	outFile.Close()
}

func Test(t string) string {
	parts := strings.Split(t, ".")
	if len(parts) == 2 {
		return parts[1]
	}
	return t
}

func (g *Generator) generateRPCStubs(methods []MethodInfo, imports map[string]string) {
	for _, m := range methods {
		g.generateRPCStub(m, imports)
	}
}

func (g *Generator) generateRPCStub(m MethodInfo, imports map[string]string) {
	entity := m.EntityName
	methodName := m.Name
	argsStruct := entity + methodName + "Args"
	respStruct := entity + methodName + "Resp"

	g.generateArgsStruct(argsStruct, m.Params, imports)
	g.generateRespStruct(respStruct, m.Results, imports)

	clientParams, clientReturn, zeroValues, returnValues := buildClientSignatures(m, imports)

	g.f.Func().
		Params(jen.Id("s").Op("*").Id(entity + "RPC")).
		Id(methodName).
		Params(clientParams...).
		Params(clientReturn...).
		BlockFunc(func(b *jen.Group) {
			buildClientBody(b, m, argsStruct, respStruct, zeroValues, returnValues, imports)
		})

	generateServerMethod(g.f, ServerMethodOptions{
		argsStruct: argsStruct,
		respStruct: respStruct,
		entity:     entity,
		m:          m,
	})
}

func (g *Generator) generateArgsStruct(name string, params []ParamInfo, imports map[string]string) {
	fields := []jen.Code{}
	for _, p := range params {
		fieldType := jenType(p.Type, imports)
		if p.IsInterface {
			fieldType = jen.Uint32()
		}
		fields = append(fields, jen.Id(p.Name).Add(fieldType))
	}
	g.f.Type().Id(name).Struct(fields...)
}

func (g *Generator) generateRespStruct(name string, results []ResultInfo, imports map[string]string) {
	fields := []jen.Code{}
	for _, r := range results {
		fields = append(fields, jen.Id(r.Name).Add(jenType(r.Type, imports)))
	}
	g.f.Type().Id(name).Struct(fields...)
}

func buildClientSignatures(m MethodInfo, imports map[string]string) (
	clientParams, clientReturn, zeroValues, returnValues []jen.Code,
) {
	if m.HasContext {
		clientParams = append(clientParams, jen.Id("ctx").Qual("context", "Context"))
	}

	for _, p := range m.Params {
		lowerName := strings.ToLower(p.Name[:1]) + p.Name[1:]
		clientParams = append(clientParams, jen.Id(lowerName).Add(jenType(p.Type, imports)))
	}

	for _, r := range m.Results {
		zeroValues = append(zeroValues, jenAddZeroValue(r.Type))
		returnValues = append(returnValues, jen.Id("resp").Op(".").Id(r.Name))
		clientReturn = append(clientReturn, jenType(r.Type, imports))
	}
	clientReturn = append(clientReturn, jen.Error())
	return
}

func buildClientBody(
	b *jen.Group,
	m MethodInfo,
	argsStruct, respStruct string,
	zeroValues, returnValues []jen.Code,
	imports map[string]string,
) {
	b.Var().Id("resp").Op("*").Id(respStruct)

	if !m.HasContext {
		b.Id("ctx").Op(":=").Id("context").Dot("Background").Call()
	}

	clientCallArgs := buildClientCallArgs(b, m, imports)

	b.Err().Op(":=").Id("s").Dot("client").Dot("Call").Call(
		jen.Id("ctx"),
		jen.Lit("Plugin."+m.Name),
		jen.Op("&").Id(argsStruct).Values(clientCallArgs...),
		jen.Op("&").Id("resp"),
	)

	b.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(append(zeroValues, jen.Err())...),
	)
	b.Return(append(returnValues, jen.Nil())...)
}

func buildClientCallArgs(b *jen.Group, m MethodInfo, imports map[string]string) []jen.Code {
	var args []jen.Code
	for _, p := range m.Params {
		lowerName := strings.ToLower(p.Name[:1]) + p.Name[1:]
		if p.IsInterface {
			server := "server" + p.Name
			brokerID := "brokerId" + p.Name

			b.Id(server).Op(":=").Op("&").Id(Test(p.Type)+"RPCServer").Values(
				jen.Id(lowerName),
				jen.Id("s").Dot("broker"),
			)
			b.Id(brokerID).Op(":=").Id("s").Dot("broker").Dot("NextId").Call()
			b.Go().Id("s").Dot("broker").Dot("AcceptAndServe").Call(
				jen.Id(brokerID),
				jen.Id(server),
			)
			args = append(args, jen.Id(p.Name).Op(":").Id(brokerID))
		} else {
			args = append(args, jen.Id(p.Name).Op(":").Id(lowerName))
		}
	}
	return args
}

func (g *Generator) getMethods(entity string) []MethodInfo {
	return parseMethodsFromPackage(g.pkg, entity)
}

func (g *Generator) addRpc() {
	for _, name := range g.entityNames {
		methods := g.getMethods(name)
		g.generateRPCStubs(methods, g.imports)
	}
}

func (g *Generator) Generate(input string, entityNames []string) error {
	if err := g.prepare(input); err != nil {
		return err
	}
	g.prepareEntities(entityNames)

	g.addImports()
	g.addPluginStructs()
	g.addRpc()

	g.output(input)

	return nil
}

func main() {
	path := os.Getenv("GOFILE")
	if path == "" {
		log.Fatal("GOFILE must be set")
	}
	if len(os.Args) < 2 {
		log.Fatal("At least one entity name must be provided")
	}

	x := Generator{}
	if err := x.Generate(path, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

type ServerMethodOptions struct {
	argsStruct string
	respStruct string
	entity     string
	m          MethodInfo
}

func generateServerMethod(f *jen.File, opts ServerMethodOptions) {
	m := opts.m
	methodName := m.Name
	entity := opts.entity

	serverParams := []jen.Code{
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("args").Op("*").Id(opts.argsStruct),
		jen.Id("resp").Op("*").Id(opts.respStruct),
	}

	resultVars := buildResultVars(m.Results)

	f.Func().
		Params(jen.Id("s").Op("*").Id(entity + "RPCServer")).
		Id(methodName).
		Params(serverParams...).
		Params(jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.Var().Err().Error()
			callArgs := buildServerCallArgs(g, m)

			addImplCall(g, m, methodName, resultVars, callArgs)
			addRespAssignment(g, m, opts.respStruct)

			g.Return(jen.Nil())
		})
}

func buildResultVars(results []ResultInfo) []jen.Code {
	var vars []jen.Code
	for _, r := range results {
		lowerName := strings.ToLower(r.Name[:1]) + r.Name[1:]
		vars = append(vars, jen.Id(lowerName))
	}
	vars = append(vars, jen.Id("err"))
	return vars
}

func buildServerCallArgs(g *jen.Group, m MethodInfo) []jen.Code {
	var args []jen.Code
	if m.HasContext {
		args = append(args, jen.Id("ctx"))
	}
	for _, p := range m.Params {
		if p.IsInterface {
			args = append(args, buildInterfaceParam(g, p))
		} else {
			args = append(args, jen.Id("args").Dot(p.Name))
		}
	}
	return args
}

func buildInterfaceParam(g *jen.Group, p ParamInfo) jen.Code {
	conn := "conn" + p.Name
	client := "client" + p.Name
	rpc := "rpc" + p.Name

	g.List(jen.Id(conn), jen.Err()).
		Op(":=").
		Id("s").Dot("broker").Dot("Dial").
		Call(jen.Id("args").Dot(p.Name))
	g.If(jen.Err().Op("!=").Nil()).
		Block(jen.Return(jen.Err()))

	g.Id(client).
		Op(":=").
		Qual(keegancsmith_rpc, "NewClient").
		Call(jen.Id(conn))

	g.Id(rpc).
		Op(":=").
		Op("&").
		Id(Test(p.Type)+"RPC").
		Values(jen.Id(client), jen.Id("s").Dot("broker"))

	return jen.Id(rpc)
}

func addImplCall(g *jen.Group, m MethodInfo, methodName string, resultVars, callArgs []jen.Code) {
	op := ":="
	if len(resultVars) == 1 {
		op = "="
	}
	g.List(resultVars...).Op(op).
		Id("s").Dot("Impl").Dot(methodName).Call(callArgs...)

	g.If(jen.Err().Op("!=").Nil()).
		Block(jen.Return(jen.Err()))
}

func addRespAssignment(g *jen.Group, m MethodInfo, respStruct string) {
	if len(m.Results) == 0 {
		g.Op("*resp = ").Id(respStruct).Values()
		return
	}

	dict := jen.Dict{}
	for _, r := range m.Results {
		lowerName := strings.ToLower(r.Name[:1]) + r.Name[1:]
		dict[jen.Id(r.Name)] = jen.Id(lowerName)
	}
	g.Op("*resp = ").Id(respStruct).Values(dict)
}
