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

package cel2sqlite

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/common/types"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types/ref"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Converter converts CEL expressions to SQLite SQL WHERE clauses
type Converter struct {
	env       *cel.Env
	columnMap map[string]ColumnInfo
	overrides []string
}

// ColumnType represents the type of a column for proper SQL generation
type ColumnType string

const (
	ColumnTypeString                ColumnType = "string"
	ColumnTypeInt                   ColumnType = "int"
	ColumnTypeBool                  ColumnType = "bool"
	ColumnTypeJSON                  ColumnType = "json"
	ColumnTypeJSONArray             ColumnType = "json_array"
	ColumnTypeOverridableField      ColumnType = "overridable"
	ColumnTypeOverridableFieldArray ColumnType = "overridable_array"
	ColumnTypeDyn                   ColumnType = "dyn"
)

// ColumnInfo contains information about a column
type ColumnInfo struct {
	SQLName string
	Type    ColumnType
}

// NewConverter creates a new converter with the given column mapping
// columnMap maps CEL variable names to ColumnInfo
func NewConverter(columnMap map[string]ColumnInfo, overrides []string) (*Converter, error) {
	envOpts := []cel.EnvOption{}

	newColumnMap := make(map[string]ColumnInfo)
	for celName, info := range columnMap {
		newColumnMap[celName] = info

		var celType *cel.Type
		switch info.Type {
		case ColumnTypeString:
			celType = cel.StringType
		case ColumnTypeInt:
			celType = cel.IntType
		case ColumnTypeBool:
			celType = cel.BoolType
		case ColumnTypeJSON:
			celType = cel.DynType
		case ColumnTypeJSONArray, ColumnTypeOverridableFieldArray:
			celType = cel.ListType(cel.DynType)
		case ColumnTypeDyn:
			celType = cel.DynType
		default:
			celType = cel.DynType
		}

		envOpts = append(envOpts, cel.Variable(celName, celType))
	}

	envOpts = append(envOpts, cel.Function("resolve",
		cel.MemberOverload("dyn_resolve_string", []*cel.Type{cel.DynType, cel.StringType}, cel.StringType,
			cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return types.String("")
			}),
		),
	))

	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		return nil, err
	}

	return &Converter{
		env:       env,
		columnMap: newColumnMap,
		overrides: overrides,
	}, nil
}

// NewConverterSimple creates a converter with simple string column mapping
// All columns are treated as dynamic type
func NewConverterSimple(columnMap map[string]string) (*Converter, error) {
	infoMap := make(map[string]ColumnInfo)
	for k, v := range columnMap {
		infoMap[k] = ColumnInfo{
			SQLName: v,
			Type:    ColumnTypeDyn,
		}
	}
	return NewConverter(infoMap, []string{})
}

func (c *Converter) Convert(input string) (string, error) {
	ast, iss := c.env.Parse(input)
	if iss.Err() != nil {
		return "", fmt.Errorf("parse error: %w", iss.Err())
	}

	checked, iss := c.env.Check(ast)
	if iss.Err() != nil {
		return "", fmt.Errorf("type check error: %w", iss.Err())
	}

	e, _ := cel.AstToCheckedExpr(checked)
	return c.exprToSQL(e.GetExpr())
}

func (c *Converter) exprToSQL(e *expr.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("nil expression")
	}

	switch exprKind := e.ExprKind.(type) {
	case *expr.Expr_ConstExpr:
		return c.literalToSQL(exprKind.ConstExpr)

	case *expr.Expr_IdentExpr:
		return c.identToSQL(exprKind.IdentExpr.Name, true)

	case *expr.Expr_CallExpr:
		return c.callToSQL(exprKind.CallExpr)

	case *expr.Expr_SelectExpr:
		return c.selectToSQL(exprKind.SelectExpr)

	case *expr.Expr_ListExpr:
		return c.listToSQL(exprKind.ListExpr)

	default:
		return "", fmt.Errorf("unsupported expr kind: %T", exprKind)
	}
}

func (c *Converter) identToSQL(name string, resolve bool) (string, error) {
	dbCol, ok := c.columnMap[name]
	if !ok {
		return "", fmt.Errorf("column not allowed: %s", name)
	}

	if resolve && (dbCol.Type == ColumnTypeOverridableField || dbCol.Type == ColumnTypeOverridableFieldArray) {
		return fmt.Sprintf("overridable_resolve(%s, '%s')", dbCol.SQLName, strings.Join(c.overrides, "|")), nil
	}

	return dbCol.SQLName, nil
}

// isJSONArrayType checks if the column is specifically a JSON array type
func (c *Converter) isJSONArrayType(e *expr.Expr) bool {
	if identExpr, ok := e.ExprKind.(*expr.Expr_IdentExpr); ok {
		if colInfo, exists := c.columnMap[identExpr.IdentExpr.Name]; exists {
			return colInfo.Type == ColumnTypeJSONArray ||
				colInfo.Type == ColumnTypeOverridableFieldArray
		}
	}
	return false
}

func (c *Converter) callToSQL(call *expr.Expr_Call) (string, error) {
	fn := call.Function
	args := call.Args

	if call.Target != nil {
		return c.handleMethod(fn, call.Target, args)
	}

	sqlArgs := make([]string, len(args))
	for i, a := range args {
		s, err := c.exprToSQL(a)
		if err != nil {
			return "", err
		}
		sqlArgs[i] = s
	}

	switch fn {
	case operators.LogicalAnd:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("AND requires 2 arguments")
		}
		return fmt.Sprintf("(%s AND %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.LogicalOr:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("OR requires 2 arguments")
		}
		return fmt.Sprintf("(%s OR %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.LogicalNot:
		if len(sqlArgs) != 1 {
			return "", fmt.Errorf("NOT requires 1 argument")
		}
		return fmt.Sprintf("(NOT %s)", sqlArgs[0]), nil

	case operators.Equals:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("equals requires 2 arguments")
		}
		return fmt.Sprintf("(%s = %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.NotEquals:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("not equals requires 2 arguments")
		}
		return fmt.Sprintf("(%s != %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.Less:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("less than requires 2 arguments")
		}
		return fmt.Sprintf("(%s < %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.LessEquals:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("less or equal requires 2 arguments")
		}
		return fmt.Sprintf("(%s <= %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.Greater:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("greater than requires 2 arguments")
		}
		return fmt.Sprintf("(%s > %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.GreaterEquals:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("greater or equal requires 2 arguments")
		}
		return fmt.Sprintf("(%s >= %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.In:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("IN requires 2 arguments")
		}

		if len(args) >= 2 && c.isJSONArrayType(args[1]) {
			rightSQL := sqlArgs[1]
			if identExpr, ok := args[1].ExprKind.(*expr.Expr_IdentExpr); ok {
				if colInfo, exists := c.columnMap[identExpr.IdentExpr.Name]; exists {
					if colInfo.Type == ColumnTypeOverridableFieldArray {
						rightSQL = fmt.Sprintf("overridable_resolve(%s, '%s')", colInfo.SQLName, strings.Join(c.overrides, "|"))
					}
				}
			}
			return fmt.Sprintf("json_array_contains(%s, %s)", rightSQL, sqlArgs[0]), nil
		}

		return fmt.Sprintf("(%s IN %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.Index:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("index requires 2 arguments")
		}
		return fmt.Sprintf("json_extract(%s, '$[' || %s || ']')", sqlArgs[0], sqlArgs[1]), nil

	case operators.Add:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("addition requires 2 arguments")
		}
		return fmt.Sprintf("(%s + %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.Subtract:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("subtraction requires 2 arguments")
		}
		return fmt.Sprintf("(%s - %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.Multiply:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("multiplication requires 2 arguments")
		}
		return fmt.Sprintf("(%s * %s)", sqlArgs[0], sqlArgs[1]), nil

	case operators.Divide:
		if len(sqlArgs) != 2 {
			return "", fmt.Errorf("division requires 2 arguments")
		}
		return fmt.Sprintf("(%s / %s)", sqlArgs[0], sqlArgs[1]), nil

	default:
		return "", fmt.Errorf("unsupported operator: %s", fn)
	}
}

func (c *Converter) handleMethod(fn string, target *expr.Expr, args []*expr.Expr) (string, error) {
	targetSQL, err := c.exprToSQL(target)
	if err != nil {
		return "", err
	}

	switch fn {
	case "resolve":
		if len(args) != 1 {
			return "", fmt.Errorf("resolve requires 1 argument")
		}
		argSQL, err := c.exprToSQL(args[0])
		if err != nil {
			return "", err
		}

		if exprKind, ok := target.ExprKind.(*expr.Expr_IdentExpr); ok {
			targetSQL, err = c.identToSQL(exprKind.IdentExpr.Name, false)
			if err != nil {
				return "", err
			}
		}

		return fmt.Sprintf("overridable_resolve(%s, %s)", targetSQL, argSQL), nil

	case "contains":
		if len(args) != 1 {
			return "", fmt.Errorf("contains requires 1 argument")
		}
		argSQL, err := c.exprToSQL(args[0])
		if err != nil {
			return "", err
		}

		if c.isJSONArrayType(target) {
			return fmt.Sprintf("json_array_contains(%s, %s)", targetSQL, argSQL), nil
		}

		return fmt.Sprintf("(%s LIKE '%%' || %s || '%%')", targetSQL, argSQL), nil

	case "startsWith":
		if len(args) != 1 {
			return "", fmt.Errorf("startsWith requires 1 argument")
		}
		argSQL, err := c.exprToSQL(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s LIKE %s || '%%')", targetSQL, argSQL), nil

	case "endsWith":
		if len(args) != 1 {
			return "", fmt.Errorf("endsWith requires 1 argument")
		}
		argSQL, err := c.exprToSQL(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s LIKE '%%' || %s)", targetSQL, argSQL), nil

	case "matches":
		if len(args) != 1 {
			return "", fmt.Errorf("matches requires 1 argument")
		}
		argSQL, err := c.exprToSQL(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("regexp_match(%s, %s)", targetSQL, argSQL), nil

	case "size":
		if c.isJSONArrayType(target) {
			return fmt.Sprintf("json_array_length(%s)", targetSQL), nil
		}
		return "", fmt.Errorf("size() is only supported for JSON arrays")

	default:
		return "", fmt.Errorf("unsupported method: %s", fn)
	}
}

func (c *Converter) selectToSQL(sel *expr.Expr_Select) (string, error) {
	// dirty hack
	if sel.Field == "raw" {
		if exprKind, ok := sel.Operand.ExprKind.(*expr.Expr_IdentExpr); ok {
			return c.identToSQL(exprKind.IdentExpr.Name, false)
		}
	}

	operandSQL, err := c.exprToSQL(sel.Operand)
	if err != nil {
		return "", err
	}

	field := sel.Field
	jsonPath := strings.ReplaceAll(field, ".", "\".")
	return fmt.Sprintf("json_extract(%s, '$.%s')", operandSQL, jsonPath), nil
}

func (c *Converter) listToSQL(list *expr.Expr_CreateList) (string, error) {
	if len(list.Elements) == 0 {
		return "()", nil
	}

	elements := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		elemSQL, err := c.exprToSQL(elem)
		if err != nil {
			return "", err
		}
		elements[i] = elemSQL
	}

	return fmt.Sprintf("(%s)", strings.Join(elements, ", ")), nil
}

func (c *Converter) literalToSQL(constant *expr.Constant) (string, error) {
	if constant == nil {
		return "NULL", nil
	}

	switch x := constant.ConstantKind.(type) {
	case *expr.Constant_StringValue:
		escaped := strings.ReplaceAll(x.StringValue, "'", "''")
		return fmt.Sprintf("'%s'", escaped), nil

	case *expr.Constant_Int64Value:
		return fmt.Sprintf("%d", x.Int64Value), nil

	case *expr.Constant_Uint64Value:
		return fmt.Sprintf("%d", x.Uint64Value), nil

	case *expr.Constant_DoubleValue:
		return fmt.Sprintf("%f", x.DoubleValue), nil

	case *expr.Constant_BoolValue:
		if x.BoolValue {
			return "1", nil
		}
		return "0", nil

	case *expr.Constant_NullValue:
		return "NULL", nil

	default:
		return "", fmt.Errorf("unsupported literal type: %T", x)
	}
}
