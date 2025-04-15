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

package build

import (
	"context"

	"gitea.plemya-x.ru/Plemya-x/ALR/internal/cliutils"
)

type ScriptViewerConfig interface {
	PagerStyle() string
}

type ScriptViewer struct {
	config ScriptViewerConfig
}

func (s *ScriptViewer) ViewScript(
	ctx context.Context,
	input *BuildInput,
	sf *ScriptFile,
	basePkg string,
) error {
	return cliutils.PromptViewScript(
		ctx,
		sf.Path,
		basePkg,
		s.config.PagerStyle(),
		input.opts.Interactive,
	)
}
