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

package gen

import (
	_ "embed"       // Пакет для встраивания содержимого файлов в бинарники Go, использовав откладку //go:embed
	"encoding/json" // Пакет для работы с JSON: декодирование и кодирование
	"errors"        // Пакет для создания и обработки ошибок
	"fmt"           // Пакет для форматированного ввода и вывода
	"io"            // Пакет для интерфейсов ввода и вывода
	"net/http"      // Пакет для HTTP-клиентов и серверов
	"text/template" // Пакет для обработки текстовых шаблонов
)

// Используем директиву //go:embed для встраивания содержимого файла шаблона в строку pipTmpl
// Встраивание файла tmpls/pip.tmpl.sh
//
//go:embed tmpls/pip.tmpl.sh
var pipTmpl string

// PipOptions содержит параметры, которые будут переданы в шаблон
type PipOptions struct {
	Name        string // Имя пакета
	Version     string // Версия пакета
	Description string // Описание пакета
}

// pypiAPIResponse представляет структуру ответа от API PyPI
type pypiAPIResponse struct {
	Info pypiInfo  `json:"info"` // Информация о пакете
	URLs []pypiURL `json:"urls"` // Список URL-адресов для загрузки пакета
}

// Метод SourceURL ищет и возвращает URL исходного distribution для пакета, если он существует
func (res pypiAPIResponse) SourceURL() (pypiURL, error) {
	for _, url := range res.URLs {
		if url.PackageType == "sdist" {
			return url, nil
		}
	}
	return pypiURL{}, errors.New("package doesn't have a source distribution")
}

// pypiInfo содержит основную информацию о пакете, такую как имя, версия и пр.
type pypiInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Summary  string `json:"summary"`
	Homepage string `json:"home_page"`
	License  string `json:"license"`
}

// pypiURL представляет информацию об одном из доступных для загрузки URL
type pypiURL struct {
	Digests     map[string]string `json:"digests"`     // Контрольные суммы для файлов
	Filename    string            `json:"filename"`    // Имя файла
	PackageType string            `json:"packagetype"` // Тип пакета (например sdist)
}

// Функция Pip загружает информацию о пакете из PyPI и использует шаблон для вывода информации
func Pip(w io.Writer, opts PipOptions) error {
	// Создаем новый шаблон с добавлением функций из FuncMap
	tmpl, err := template.New("pip").
		Funcs(funcs).
		Parse(pipTmpl)
	if err != nil {
		return err
	}

	// Формируем URL для запроса к PyPI на основании имени и версии пакета
	url := fmt.Sprintf(
		"https://pypi.org/pypi/%s/%s/json",
		opts.Name,
		opts.Version,
	)

	// Выполняем HTTP GET запрос к PyPI
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close() // Закрываем тело ответа после завершения работы
	if res.StatusCode != 200 {
		return fmt.Errorf("pypi: %s", res.Status)
	}

	// Раскодируем ответ JSON от PyPI в структуру pypiAPIResponse
	var resp pypiAPIResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return err
	}

	// Если в opts указано описание, используем его вместо описания из PyPI
	if opts.Description != "" {
		resp.Info.Summary = opts.Description
	}

	// Выполняем шаблон с использованием данных из resp и записываем результат в w
	return tmpl.Execute(w, resp)
}
