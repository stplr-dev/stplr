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

//go:build e2e

package e2etests_test

import (
	"testing"

	"go.alt-gnome.ru/capytest"
)

func TestE2EAppstreamDescriptionI18n(t *testing.T) {
	t.Parallel()

	t.Run("stplr info returns correct description", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		expected := "<p>\n      Lunacy — идеальный графический редактор для UX/UI и веб-дизайна." +
			" Он объединяет лучшие функции всех дизайнерских приложений, помогая оптимизировать" +
			" рабочий процесс и сосредоточиться на творчестве. Пользуйтесь встроенной библиотекой" +
			" графики, мощными инструментами на основе ИИ и совместно работайте над проектами с" +
			" командой на нескольких платформах одновременно!\n    </p>" +
			"<p>Давайте посмотрим, что внутри:</p>" +
			"<ul>" +
			"<li>Совместная работа в реальном времени с комментариями, стикерами и аудиосообщениями</li>" +
			"<li>1 500 000 иконок, фотографий и иллюстраций</li>" +
			"<li>Инструменты ИИ для рутинных задач: удаление фона, улучшение качества изображений," +
			" генерация текста и аватаров</li>" +
			"<li>Поддержка форматов Figma и Sketch</li>" +
			"<li>Облачное хранилище</li>" +
			"<li>Приватный офлайн-режим</li>" +
			"<li>Низкие системные требования</li>" +
			"<li>Поддержка 26 языков</li>" +
			"</ul>" +
			"<p>И многое другое!</p>" +
			"<p>Присоединяйтесь к нашему растущему сообществу дизайнеров!</p>"

		r.Command(
			"stplr", "search",
			"-q", "name == 'test-appstream-metainfo'",
			"-f", "{{ if and .AppStream .AppStream.Description }}{{ .AppStream.Description | localized }}{{ end }}",
		).
			WithEnv("LANG", "ru").
			ExpectStdoutEqual(expected).
			ExpectSuccess().
			Run(t)
	}))
}
