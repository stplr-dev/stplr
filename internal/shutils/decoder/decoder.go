// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package decoder

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slices"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/pkg/distro"
)

var ErrNotPointerToStruct = errors.New("val must be a pointer to a struct")

type VarNotFoundError struct {
	name string
}

func (nfe VarNotFoundError) Error() string {
	return "required variable '" + nfe.name + "' could not be found"
}

type InvalidTypeError struct {
	name    string
	vartype string
	exptype string
}

func (ite InvalidTypeError) Error() string {
	return fmt.Sprintf("variable '%s' is of type %s, but %s is expected", ite.name, ite.vartype, ite.exptype)
}

// Decoder provides methods for decoding variable values
type Decoder struct {
	info   *distro.OSRelease
	Runner *interp.Runner
	// Enable distro overrides (true by default)
	Overrides bool
	// Enable using like distros for overrides
	LikeDistros bool

	OverridesOpts *overrides.Opts
}

// New creates a new variable decoder
func New(info *distro.OSRelease, runner *interp.Runner) *Decoder {
	return &Decoder{
		info,
		runner,
		true,
		len(info.Like) > 0,
		overrides.DefaultOpts,
	}
}

// DecodeVar decodes a variable to val using reflection.
// Structs should use the "sh" struct tag.
func (d *Decoder) DecodeVar(name string, val any) error {
	origType := reflect.TypeOf(val).Elem()
	isOverridableField := strings.Contains(origType.String(), "OverridableField[")

	if !isOverridableField {
		variable := d.getVarNoOverrides(name)
		if variable == nil {
			return VarNotFoundError{name}
		}

		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           val, // передаем указатель на новое значение
			TagName:          "sh",
			DecodeHook: mapstructure.DecodeHookFuncValue(func(from, to reflect.Value) (interface{}, error) {
				if from.Kind() == reflect.Slice && to.Kind() == reflect.String {
					s, ok := from.Interface().([]string)
					if ok && len(s) == 1 {
						return s[0], nil
					}
				}
				return from.Interface(), nil
			}),
		})
		if err != nil {
			return err
		}

		switch variable.Kind {
		case expand.Indexed:
			return dec.Decode(variable.List)
		case expand.Associative:
			return dec.Decode(variable.Map)
		default:
			return dec.Decode(variable.Str)
		}
	} else {
		vars := d.getVarsByPrefix(name)

		if len(vars) == 0 {
			return VarNotFoundError{name}
		}

		reflectVal := reflect.ValueOf(val)
		overridableVal := reflect.ValueOf(val).Elem()

		dataField := overridableVal.FieldByName("data")
		if !dataField.IsValid() {
			return fmt.Errorf("data field not found in OverridableField")
		}
		mapType := dataField.Type() // map[string]T
		elemType := mapType.Elem()  // T

		var overridablePtr reflect.Value
		if reflectVal.Kind() == reflect.Ptr {
			overridablePtr = reflectVal
		} else {
			if !reflectVal.CanAddr() {
				return fmt.Errorf("OverridableField value is not addressable")
			}
			overridablePtr = reflectVal.Addr()
		}

		setValue := overridablePtr.MethodByName("Set")
		if !setValue.IsValid() {
			return fmt.Errorf("method Set not found on OverridableField")
		}

		for _, v := range vars {
			varName := v.Name

			key := strings.TrimPrefix(strings.TrimPrefix(varName, name), "_")
			newVal := reflect.New(elemType)

			if err := d.DecodeVar(varName, newVal.Interface()); err != nil {
				return err
			}

			keyValue := reflect.ValueOf(key)
			setValue.Call([]reflect.Value{keyValue, newVal.Elem()})
		}

		resolveValue := overridablePtr.MethodByName("Resolve")
		if !resolveValue.IsValid() {
			return fmt.Errorf("method Resolve not found on OverridableField")
		}

		names, err := overrides.Resolve(d.info, d.OverridesOpts)
		if err != nil {
			return err
		}

		resolveValue.Call([]reflect.Value{reflect.ValueOf(names)})
		return nil
	}
}

// DecodeVars decodes all variables to val using reflection.
// Structs should use the "sh" struct tag.
func (d *Decoder) DecodeVars(val any) error {
	valKind := reflect.TypeOf(val).Kind()
	if valKind != reflect.Pointer {
		return ErrNotPointerToStruct
	} else {
		elemKind := reflect.TypeOf(val).Elem().Kind()
		if elemKind != reflect.Struct {
			return ErrNotPointerToStruct
		}
	}

	rVal := reflect.ValueOf(val).Elem()
	return d.decodeStruct(rVal)
}

func (d *Decoder) decodeStruct(rVal reflect.Value) error {
	for i := 0; i < rVal.NumField(); i++ {
		field := rVal.Field(i)
		fieldType := rVal.Type().Field(i)

		// Пропускаем неэкспортируемые поля
		if !fieldType.IsExported() {
			continue
		}

		// Обрабатываем встроенные поля рекурсивно
		if fieldType.Anonymous {
			if field.Kind() == reflect.Struct {
				if err := d.decodeStruct(field); err != nil {
					return err
				}
			}
			continue
		}

		name := fieldType.Name
		tag := fieldType.Tag.Get("sh")
		required := false
		if tag != "" {
			if strings.Contains(tag, ",") {
				splitTag := strings.Split(tag, ",")
				name = splitTag[0]

				if len(splitTag) > 1 {
					if slices.Contains(splitTag, "required") {
						required = true
					}
				}
			} else {
				name = tag
			}
		}

		newVal := reflect.New(field.Type())
		err := d.DecodeVar(name, newVal.Interface())
		if _, ok := err.(VarNotFoundError); ok && !required {
			continue
		} else if err != nil {
			return err
		}

		field.Set(newVal.Elem())
	}
	return nil
}

type (
	ScriptFunc             func(ctx context.Context, opts ...interp.RunnerOption) error
	ScriptFuncWithSubshell func(ctx context.Context, opts ...interp.RunnerOption) (*interp.Runner, error)
)

// GetFunc returns a function corresponding to a bash function
// with the given name
func (d *Decoder) GetFunc(name string) (ScriptFunc, bool) {
	return d.GetFuncP(name, nil)
}

type PrepareFunc func(context.Context, *interp.Runner) error

func (d *Decoder) GetFuncP(name string, prepare PrepareFunc) (ScriptFunc, bool) {
	fn := d.getFunc(name)
	if fn == nil {
		return nil, false
	}

	return func(ctx context.Context, opts ...interp.RunnerOption) error {
		sub := d.Runner.Subshell()
		for _, opt := range opts {
			err := opt(sub)
			if err != nil {
				return err
			}
		}
		if prepare != nil {
			if err := prepare(ctx, sub); err != nil {
				return err
			}
		}
		return sub.Run(ctx, fn)
	}, true
}

func (d *Decoder) GetFuncWithSubshell(name string) (ScriptFuncWithSubshell, bool) {
	fn := d.getFunc(name)
	if fn == nil {
		return nil, false
	}

	return func(ctx context.Context, opts ...interp.RunnerOption) (*interp.Runner, error) {
		sub := d.Runner.Subshell()
		for _, opt := range opts {
			err := opt(sub)
			if err != nil {
				return nil, err
			}
		}
		return sub, sub.Run(ctx, fn)
	}, true
}

func (d *Decoder) getFunc(name string) *syntax.Stmt {
	names, err := overrides.Resolve(d.info, overrides.DefaultOpts.WithName(name))
	if err != nil {
		return nil
	}

	for _, fnName := range names {
		fn, ok := d.Runner.Funcs[fnName]
		if ok {
			return fn
		}
	}
	return nil
}

func (d *Decoder) getVarNoOverrides(name string) *expand.Variable {
	val, ok := d.Runner.Vars[name]
	if ok {
		// Resolve nameref variables
		_, resolved := val.Resolve(expand.FuncEnviron(func(s string) string {
			if val, ok := d.Runner.Vars[s]; ok {
				return val.String()
			}
			return ""
		}))
		val = resolved

		return &val
	}
	return nil
}

type vars struct {
	Name  string
	Value *expand.Variable
}

var skipSubPrefixesMap = map[string][]string{
	"auto_req":  {"auto_req_skiplist"},
	"auto_prov": {"auto_prov_skiplist"},
	"nonfree":   {"nonfree_msg", "nonfree_msgfile", "nonfree_url"},
}

func (d *Decoder) getVarsByPrefix(prefix string) []*vars {
	result := make([]*vars, 0)
	skipSubPrefixes := skipSubPrefixesMap[prefix]

	for name, val := range d.Runner.Vars {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		shouldSkip := false
		for _, skipPrefix := range skipSubPrefixes {
			if strings.HasPrefix(name, skipPrefix) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}
		result = append(result, &vars{name, &val})
	}
	return result
}

func IsTruthy(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "yes" || value == "1"
}
