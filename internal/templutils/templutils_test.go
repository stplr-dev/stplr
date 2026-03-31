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

package templutils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.alt-gnome.ru/x/appstream"
)

func TestLocalizedText(t *testing.T) {
	m := appstream.LocalizedMap{
		{Lang: "ru", Value: "Привет"},
		{Lang: "en", Value: "Hello"},
		{Lang: "", Value: "Default"},
	}

	t.Run("exact lang match", func(t *testing.T) {
		assert.Equal(t, "Привет", localizedText(m, "ru"))
	})

	t.Run("strips encoding suffix", func(t *testing.T) {
		assert.Equal(t, "Привет", localizedText(m, "ru_RU.UTF-8"))
	})

	t.Run("strips region suffix", func(t *testing.T) {
		assert.Equal(t, "Привет", localizedText(m, "ru_RU"))
	})

	t.Run("fallback to next lang", func(t *testing.T) {
		assert.Equal(t, "Hello", localizedText(m, "fr", "en"))
	})

	t.Run("empty lang matches default entry", func(t *testing.T) {
		assert.Equal(t, "Default", localizedText(m, ""))
	})

	t.Run("no match returns empty string", func(t *testing.T) {
		assert.Equal(t, "", localizedText(m, "ja"))
	})

	t.Run("empty map returns empty string", func(t *testing.T) {
		assert.Equal(t, "", localizedText(appstream.LocalizedMap{}, "en"))
	})

	t.Run("no langs uses LANG env fallback order", func(t *testing.T) {
		t.Setenv("LANG", "ru_RU.UTF-8")
		assert.Equal(t, "Привет", localizedText(m))
	})

	t.Run("no langs with unmatched LANG falls back to en", func(t *testing.T) {
		t.Setenv("LANG", "ja_JP.UTF-8")
		assert.Equal(t, "Hello", localizedText(m))
	})
}

func TestIndent(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		assert.Equal(t, "  hello", indent("hello"))
	})

	t.Run("multiline", func(t *testing.T) {
		assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc"))
	})

	t.Run("empty lines not indented", func(t *testing.T) {
		assert.Equal(t, "  a\n\n  b", indent("a\n\nb"))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", indent(""))
	})
}

func TestNewPackageTemplate(t *testing.T) {
	t.Run("returns non-nil template", func(t *testing.T) {
		tmpl := NewPackageTemplate()
		require.NotNil(t, tmpl)
	})

	t.Run("localized func available", func(t *testing.T) {
		tmpl := NewPackageTemplate()
		_, err := tmpl.Parse(`{{localized . "en"}}`)
		require.NoError(t, err)

		m := appstream.LocalizedMap{{Lang: "en", Value: "Hello"}}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, m)
		require.NoError(t, err)
		assert.Equal(t, "Hello", buf.String())
	})

	t.Run("indent func available", func(t *testing.T) {
		tmpl := NewPackageTemplate()
		_, err := tmpl.Parse(`{{indent .}}`)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, "line1\nline2")
		require.NoError(t, err)
		assert.Equal(t, "  line1\n  line2", buf.String())
	})
}
