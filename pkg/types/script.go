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

package types

import (
	"bytes"
	"encoding/gob"

	"mvdan.cc/sh/v3/syntax"
	"mvdan.cc/sh/v3/syntax/typedjson"
)

type ScriptFile struct {
	File *syntax.File
	Path string
}

func (s *ScriptFile) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(s.Path); err != nil {
		return nil, err
	}
	var fileBuf bytes.Buffer
	if err := typedjson.Encode(&fileBuf, s.File); err != nil {
		return nil, err
	}
	fileData := fileBuf.Bytes()
	if err := enc.Encode(fileData); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ScriptFile) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&s.Path); err != nil {
		return err
	}
	var fileData []byte
	if err := dec.Decode(&fileData); err != nil {
		return err
	}
	fileReader := bytes.NewReader(fileData)
	file, err := typedjson.Decode(fileReader)
	if err != nil {
		return err
	}
	s.File = file.(*syntax.File)
	return nil
}
