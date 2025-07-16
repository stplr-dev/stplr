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

package staplerfile

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

type OverridableField[T any] struct {
	data map[string]T
	// It can't be a pointer
	//
	// See https://gitea.com/xorm/xorm/issues/2431
	resolved T
}

func (f *OverridableField[T]) Set(key string, value T) {
	if f.data == nil {
		f.data = make(map[string]T)
	}
	f.data[key] = value
}

func (f *OverridableField[T]) Get(key string) T {
	if f.data == nil {
		f.data = make(map[string]T)
	}
	return f.data[key]
}

func (f *OverridableField[T]) Has(key string) (T, bool) {
	if f.data == nil {
		f.data = make(map[string]T)
	}
	val, ok := f.data[key]
	return val, ok
}

func (f *OverridableField[T]) SetResolved(value T) {
	f.resolved = value
}

func (f *OverridableField[T]) Resolved() T {
	return f.resolved
}

func (f *OverridableField[T]) All() map[string]T {
	return f.data
}

func (o *OverridableField[T]) Resolve(overrides []string) {
	for _, override := range overrides {
		if v, ok := o.Has(override); ok {
			o.SetResolved(v)
			return
		}
	}
}

// Database serialization (JSON)
func (f *OverridableField[T]) ToDB() ([]byte, error) {
	var data map[string]T

	if f.data == nil {
		data = make(map[string]T)
	} else {
		data = f.data
	}

	return json.Marshal(data)
}

func (f *OverridableField[T]) FromDB(data []byte) error {
	if len(data) == 0 {
		*f = OverridableField[T]{data: make(map[string]T)}
		return nil
	}

	var temp map[string]T
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp == nil {
		temp = make(map[string]T)
	}

	*f = OverridableField[T]{data: temp}
	return nil
}

// Gob serialization
type overridableFieldGobPayload[T any] struct {
	Data     map[string]T
	Resolved T
}

func (f *OverridableField[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	payload := overridableFieldGobPayload[T]{
		Data:     f.data,
		Resolved: f.resolved,
	}

	if err := enc.Encode(payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f *OverridableField[T]) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	var payload overridableFieldGobPayload[T]
	if err := dec.Decode(&payload); err != nil {
		return err
	}

	f.data = payload.Data
	f.resolved = payload.Resolved
	return nil
}

type overridableFieldJSONPayload[T any] struct {
	Resolved *T           `json:"resolved,omitempty,omitzero"`
	Data     map[string]T `json:"overrides,omitempty,omitzero"`
}

func (f OverridableField[T]) MarshalJSON() ([]byte, error) {
	data := make(map[string]T)

	for k, v := range f.data {
		if k == "" {
			data["default"] = v
		} else {
			data[k] = v
		}
	}

	payload := overridableFieldJSONPayload[T]{
		Data:     data,
		Resolved: &f.resolved,
	}

	return json.Marshal(payload)
}

func (f *OverridableField[T]) UnmarshalJSON(data []byte) error {
	var payload overridableFieldJSONPayload[T]
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.Data == nil {
		payload.Data = make(map[string]T)
	}

	f.data = payload.Data
	if payload.Resolved != nil {
		f.resolved = *payload.Resolved
	}

	return nil
}

func OverridableFromMap[T any](data map[string]T) OverridableField[T] {
	if data == nil {
		data = make(map[string]T)
	}
	return OverridableField[T]{
		data: data,
	}
}
