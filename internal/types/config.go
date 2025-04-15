// This file was originally part of the project "LURE - Linux User REpository", created by Elara Musayelyan.
// It has been modified as part of "ALR - Any Linux Repository" by Евгений Храмов.
//
// ALR - Any Linux Repository
// Copyright (C) 2025 Евгений Храмов
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

// Config represents the ALR configuration file
type Config struct {
	RootCmd          string   `toml:"rootCmd" env:"ALR_ROOT_CMD"`
	PagerStyle       string   `toml:"pagerStyle" env:"ALR_PAGER_STYLE"`
	IgnorePkgUpdates []string `toml:"ignorePkgUpdates"`
	Repos            []Repo   `toml:"repo"`
	AutoPull         bool     `toml:"autoPull" env:"ALR_AUTOPULL"`
	LogLevel         string   `toml:"logLevel" env:"ALR_LOG_LEVEL"`
}

// Repo represents a ALR repo within a configuration file
type Repo struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}
