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

//go:generate bash -c "go run ../../generators/staplerfile-package && cd ../.. && make update-license"

package staplerfile

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"go.stplr.dev/stplr/internal/shutils/decoder"
)

type PackageNames struct {
	BasePkgName string   `sh:"basepkg_name"`
	Names       []string `sh:"name"`
}

func ParseNames(dec *decoder.Decoder) (*PackageNames, error) {
	var pkgs PackageNames
	err := dec.DecodeVars(&pkgs)
	if err != nil {
		return nil, fmt.Errorf("fail parse names: %w", err)
	}
	return &pkgs, nil
}

type Package struct {
	Repository  string `xorm:"pk 'repository'" json:"repository"`
	Name        string `xorm:"pk 'name'" json:"name"`
	BasePkgName string `xorm:"notnull 'basepkg_name'" json:"basepkg_name"`

	Version       string   `sh:"version" xorm:"notnull 'version'" json:"version"`
	Release       int      `sh:"release" xorm:"notnull 'release'" json:"release"`
	Epoch         uint     `sh:"epoch" xorm:"'epoch'" json:"epoch"`
	Architectures []string `sh:"architectures" xorm:"json 'architectures'" json:"architectures"`
	Licenses      []string `sh:"license" xorm:"json 'licenses'" json:"license"`
	Provides      []string `sh:"provides" xorm:"json 'provides'" json:"provides"`
	Conflicts     []string `sh:"conflicts" xorm:"json 'conflicts'" json:"conflicts"`
	Replaces      []string `sh:"replaces" xorm:"json 'replaces'" json:"replaces"`

	Summary          OverridableField[string]   `sh:"summary" xorm:"'summary'" json:"summary"`
	Description      OverridableField[string]   `sh:"desc" xorm:"'description'" json:"description"`
	Group            OverridableField[string]   `sh:"group" xorm:"'group_name'" json:"group"`
	Homepage         OverridableField[string]   `sh:"homepage" xorm:"'homepage'" json:"homepage"`
	Maintainer       OverridableField[string]   `sh:"maintainer" xorm:"'maintainer'" json:"maintainer"`
	Depends          OverridableField[[]string] `sh:"deps" xorm:"'depends'" json:"deps"`
	BuildDepends     OverridableField[[]string] `sh:"build_deps" xorm:"'builddepends'" json:"build_deps"`
	OptDepends       OverridableField[[]string] `sh:"opt_deps" xorm:"'optdepends'" json:"opt_deps,omitempty"`
	Sources          OverridableField[[]string] `sh:"sources" xorm:"-" json:"sources"`
	Checksums        OverridableField[[]string] `sh:"checksums" xorm:"-" json:"checksums,omitempty"`
	Backup           OverridableField[[]string] `sh:"backup" xorm:"-" json:"backup"`
	Scripts          OverridableField[Scripts]  `sh:"scripts" xorm:"-" json:"scripts,omitempty"`
	AutoReq          OverridableField[[]string] `sh:"auto_req" xorm:"-" json:"auto_req"`
	AutoProv         OverridableField[[]string] `sh:"auto_prov" xorm:"-" json:"auto_prov"`
	AutoReqSkipList  OverridableField[[]string] `sh:"auto_req_skiplist" xorm:"-" json:"auto_req_skiplist,omitempty"`
	AutoProvSkipList OverridableField[[]string] `sh:"auto_prov_skiplist" xorm:"-" json:"auto_prov_skiplist,omitempty"`

	FireJailed       OverridableField[bool]              `sh:"firejailed" xorm:"-" json:"firejailed"`
	FireJailProfiles OverridableField[map[string]string] `sh:"firejail_profiles" xorm:"-" json:"firejail_profiles,omitempty"`
}

type Scripts struct {
	PreInstall  string `sh:"preinstall"`
	PostInstall string `sh:"postinstall"`
	PreRemove   string `sh:"preremove"`
	PostRemove  string `sh:"postremove"`
	PreUpgrade  string `sh:"preupgrade"`
	PostUpgrade string `sh:"postupgrade"`
	PreTrans    string `sh:"pretrans"`
	PostTrans   string `sh:"posttrans"`
}

func (p Package) MarshalJSONWithOptions(includeOverrides bool) ([]byte, error) {
	// Сначала сериализуем обычным способом для получения базовой структуры
	type PackageAlias Package
	baseData, err := json.Marshal(PackageAlias(p))
	if err != nil {
		return nil, err
	}

	// Десериализуем в map для модификации
	var result map[string]json.RawMessage
	if err := json.Unmarshal(baseData, &result); err != nil {
		return nil, err
	}

	// Теперь заменяем OverridableField поля
	v := reflect.ValueOf(p)
	t := reflect.TypeOf(p)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldName := jsonTag
		if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
			fieldName = jsonTag[:commaIdx]
		}

		if field.Type().Name() == "OverridableField" ||
			(field.Type().Kind() == reflect.Struct &&
				strings.Contains(field.Type().String(), "OverridableField")) {

			fieldPtr := field.Addr()

			resolvedMethod := fieldPtr.MethodByName("Resolved")
			if resolvedMethod.IsValid() {
				resolved := resolvedMethod.Call(nil)[0]

				fieldData := map[string]interface{}{
					"resolved": resolved.Interface(),
				}

				if includeOverrides {
					allMethod := field.MethodByName("All")
					if allMethod.IsValid() {
						overrides := allMethod.Call(nil)[0]
						if !overrides.IsNil() && overrides.Len() > 0 {
							fieldData["overrides"] = overrides.Interface()
						}
					}
				}

				fieldJSON, err := json.Marshal(fieldData)
				if err != nil {
					return nil, err
				}
				result[fieldName] = json.RawMessage(fieldJSON)
			}
		}
	}

	return json.Marshal(result)
}
