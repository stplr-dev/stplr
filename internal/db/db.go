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

package db

import (
	"context"
	"errors"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"

	"go.stplr.dev/stplr/pkg/staplerfile"

	"go.stplr.dev/stplr/internal/config"
)

const CurrentVersion = 5

type Version struct {
	Version int `xorm:"'version'"`
}

type Config interface {
	GetPaths() *config.Paths
}

type Database struct {
	engine *xorm.Engine
	config Config
}

func New(config Config) *Database {
	return &Database{
		config: config,
	}
}

func (d *Database) Connect() error {
	dsn := d.config.GetPaths().DBPath
	engine, err := xorm.NewEngine("sqlite", dsn)
	// engine.SetLogLevel(log.LOG_DEBUG)
	// engine.ShowSQL(true)
	if err != nil {
		return err
	}
	d.engine = engine
	return nil
}

func (d *Database) Init(ctx context.Context) error {
	if err := d.Connect(); err != nil {
		return err
	}

	if err := d.sync(); err != nil {
		return err
	}
	ver, ok := d.GetVersion(ctx)

	switch {
	case !ok:
		return d.addVersion()
	case ver != CurrentVersion:
		return errors.New("incorrect db version")
	}
	return nil
}

func (d *Database) GetVersion(ctx context.Context) (int, bool) {
	var v Version
	has, err := d.engine.Get(&v)
	if err != nil || !has {
		return 0, false
	}
	return v.Version, true
}

func (d *Database) addVersion() error {
	_, err := d.engine.Insert(&Version{Version: CurrentVersion})
	return err
}

func (d *Database) sync() error {
	return d.engine.Sync(new(staplerfile.Package), new(Version))
}

func (d *Database) reset() error {
	return d.engine.DropTables(new(staplerfile.Package), new(Version))
}

func (d *Database) InsertPackage(ctx context.Context, pkg staplerfile.Package) error {
	session := d.engine.Context(ctx)

	affected, err := session.Where("name = ? AND repository = ?", pkg.Name, pkg.Repository).Update(&pkg)
	if err != nil {
		return err
	}

	if affected == 0 {
		_, err = session.Insert(&pkg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) GetPkgs(_ context.Context, where string, args ...any) ([]staplerfile.Package, error) {
	var pkgs []staplerfile.Package
	err := d.engine.Where(where, args...).Find(&pkgs)
	return pkgs, err
}

func (d *Database) GetPkg(where string, args ...any) (*staplerfile.Package, error) {
	var pkg staplerfile.Package
	has, err := d.engine.Where(where, args...).Get(&pkg)
	if err != nil || !has {
		return nil, err
	}
	return &pkg, nil
}

func (d *Database) DeletePkgs(_ context.Context, where string, args ...any) error {
	_, err := d.engine.Where(where, args...).Delete(&staplerfile.Package{})
	return err
}

func (d *Database) IsEmpty() bool {
	count, err := d.engine.Count(new(staplerfile.Package))
	return err != nil || count == 0
}

func (d *Database) Close() error {
	return d.engine.Close()
}
