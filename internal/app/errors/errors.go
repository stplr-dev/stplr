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

package errors

import (
	"fmt"
)

type I18nError struct {
	Cause   error
	Message string
}

func (e *I18nError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e I18nError) Unwrap() error {
	return e.Cause
}

func NewI18nError(message string) error {
	return &I18nError{
		Message: message,
	}
}

func WrapIntoI18nError(err error, msg string) error {
	return &I18nError{
		err,
		msg,
	}
}
