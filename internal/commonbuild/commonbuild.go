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

package commonbuild

import (
	"bytes"
	"encoding/gob"

	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"
)

type BuildInput struct {
	BasePkgName string
	Opts        *types.BuildOpts
	Info_       *distro.OSRelease
	PkgFormat_  string
	Script      string
	Repository_ string
	Packages_   []string
}

func (bi *BuildInput) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	if err := encoder.Encode(bi.BasePkgName); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Opts); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Info_); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.PkgFormat_); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Script); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Repository_); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Packages_); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (bi *BuildInput) GobDecode(data []byte) error {
	r := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(r)

	if err := decoder.Decode(&bi.BasePkgName); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Opts); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Info_); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.PkgFormat_); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Script); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Repository_); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Packages_); err != nil {
		return err
	}

	return nil
}

func (b *BuildInput) Repository() string {
	return b.Repository_
}

func (b *BuildInput) BuildOpts() *types.BuildOpts {
	return b.Opts
}

func (b *BuildInput) OSRelease() *distro.OSRelease {
	return b.Info_
}

func (b *BuildInput) PkgFormat() string {
	return b.PkgFormat_
}

func (b *BuildInput) Packages() []string {
	return b.Packages_
}
