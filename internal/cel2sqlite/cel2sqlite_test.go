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

package cel2sqlite_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/cel2sqlite"
)

func TestConverterBasicComparisons(t *testing.T) {
	columnMap := map[string]string{
		"age":    "users.age",
		"name":   "users.name",
		"status": "users.status",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "simple equality",
			celExpr:  "age == 25",
			expected: "(users.age = 25)",
		},
		{
			name:     "string equality",
			celExpr:  "name == 'John'",
			expected: "(users.name = 'John')",
		},
		{
			name:     "greater than",
			celExpr:  "age > 18",
			expected: "(users.age > 18)",
		},
		{
			name:     "less than or equal",
			celExpr:  "age <= 65",
			expected: "(users.age <= 65)",
		},
		{
			name:     "not equal",
			celExpr:  "status != 'inactive'",
			expected: "(users.status != 'inactive')",
		},
		{
			name:     "less than",
			celExpr:  "age < 30",
			expected: "(users.age < 30)",
		},
		{
			name:     "greater than or equal",
			celExpr:  "age >= 21",
			expected: "(users.age >= 21)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterLogicalOperators(t *testing.T) {
	columnMap := map[string]string{
		"age":    "users.age",
		"status": "users.status",
		"active": "users.active",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "AND operator",
			celExpr:  "age > 18 && status == 'active'",
			expected: "((users.age > 18) AND (users.status = 'active'))",
		},
		{
			name:     "OR operator",
			celExpr:  "age < 18 || age > 65",
			expected: "((users.age < 18) OR (users.age > 65))",
		},
		{
			name:     "NOT operator",
			celExpr:  "!active",
			expected: "(NOT users.active)",
		},
		{
			name:     "complex expression",
			celExpr:  "(age >= 18 && age <= 65) && status == 'active'",
			expected: "(((users.age >= 18) AND (users.age <= 65)) AND (users.status = 'active'))",
		},
		{
			name:     "multiple OR conditions",
			celExpr:  "status == 'active' || status == 'pending' || status == 'review'",
			expected: "(((users.status = 'active') OR (users.status = 'pending')) OR (users.status = 'review'))",
		},
		{
			name:     "NOT with comparison",
			celExpr:  "!(age < 18)",
			expected: "(NOT (users.age < 18))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterStringMethods(t *testing.T) {
	columnMap := map[string]string{
		"email": "users.email",
		"name":  "users.name",
		"phone": "users.phone",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "contains",
			celExpr:  "email.contains('@gmail.com')",
			expected: "(users.email LIKE '%' || '@gmail.com' || '%')",
		},
		{
			name:     "startsWith",
			celExpr:  "name.startsWith('John')",
			expected: "(users.name LIKE 'John' || '%')",
		},
		{
			name:     "endsWith",
			celExpr:  "email.endsWith('.com')",
			expected: "(users.email LIKE '%' || '.com')",
		},
		{
			name:     "contains with special characters",
			celExpr:  "email.contains('@')",
			expected: "(users.email LIKE '%' || '@' || '%')",
		},
		{
			name:     "combined string methods",
			celExpr:  "email.contains('@gmail.com') && name.startsWith('A')",
			expected: "((users.email LIKE '%' || '@gmail.com' || '%') AND (users.name LIKE 'A' || '%'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterInOperator(t *testing.T) {
	columnMap := map[string]string{
		"status": "users.status",
		"id":     "users.id",
		"role":   "users.role",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "IN with strings",
			celExpr:  "status in ['active', 'pending']",
			expected: "(users.status IN ('active', 'pending'))",
		},
		{
			name:     "IN with numbers",
			celExpr:  "id in [1, 2, 3]",
			expected: "(users.id IN (1, 2, 3))",
		},
		{
			name:     "IN with single value",
			celExpr:  "status in ['active']",
			expected: "(users.status IN ('active'))",
		},
		{
			name:     "IN with multiple strings",
			celExpr:  "role in ['admin', 'moderator', 'user', 'guest']",
			expected: "(users.role IN ('admin', 'moderator', 'user', 'guest'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterArithmetic(t *testing.T) {
	columnMap := map[string]string{
		"price":    "products.price",
		"discount": "products.discount",
		"quantity": "products.quantity",
		"tax":      "products.tax",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "addition",
			celExpr:  "price + discount > 100",
			expected: "((products.price + products.discount) > 100)",
		},
		{
			name:     "subtraction",
			celExpr:  "price - discount < 50",
			expected: "((products.price - products.discount) < 50)",
		},
		{
			name:     "multiplication",
			celExpr:  "price * quantity > 1000",
			expected: "((products.price * products.quantity) > 1000)",
		},
		{
			name:     "division",
			celExpr:  "price / 2 >= 25",
			expected: "((products.price / 2) >= 25)",
		},
		{
			name:     "complex arithmetic",
			celExpr:  "(price - discount) * quantity > 500",
			expected: "(((products.price - products.discount) * products.quantity) > 500)",
		},
		{
			name:     "multiple operations",
			celExpr:  "price * quantity + tax < 1000",
			expected: "(((products.price * products.quantity) + products.tax) < 1000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterDataTypes(t *testing.T) {
	columnMap := map[string]string{
		"count":   "items.count",
		"price":   "items.price",
		"name":    "items.name",
		"active":  "items.active",
		"deleted": "items.deleted",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "integer comparison",
			celExpr:  "count == 42",
			expected: "(items.count = 42)",
		},
		{
			name:     "boolean true",
			celExpr:  "active == true",
			expected: "(items.active = 1)",
		},
		{
			name:     "boolean false",
			celExpr:  "deleted == false",
			expected: "(items.deleted = 0)",
		},
		{
			name:     "string with quotes",
			celExpr:  "name == \"Product's Name\"",
			expected: "(items.name = 'Product''s Name')",
		},
		{
			name:     "negative number",
			celExpr:  "count < -5",
			expected: "(items.count < -5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterComplexExpressions(t *testing.T) {
	columnMap := map[string]string{
		"age":    "users.age",
		"status": "users.status",
		"email":  "users.email",
		"role":   "users.role",
		"salary": "users.salary",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "complex AND OR combination",
			celExpr:  "(age >= 18 && age <= 65) && (status == 'active' || status == 'pending')",
			expected: "(((users.age >= 18) AND (users.age <= 65)) AND ((users.status = 'active') OR (users.status = 'pending')))",
		},
		{
			name:     "NOT with complex condition",
			celExpr:  "!(age < 18 || age > 65) && status == 'active'",
			expected: "((NOT ((users.age < 18) OR (users.age > 65))) AND (users.status = 'active'))",
		},
		{
			name:     "string methods with logical operators",
			celExpr:  "email.contains('@gmail.com') && age > 18 && status in ['active', 'verified']",
			expected: "(((users.email LIKE '%' || '@gmail.com' || '%') AND (users.age > 18)) AND (users.status IN ('active', 'verified')))",
		},
		{
			name:     "arithmetic with comparisons",
			celExpr:  "salary * 12 >= 50000 && (status == 'active' || role == 'admin')",
			expected: "(((users.salary * 12) >= 50000) AND ((users.status = 'active') OR (users.role = 'admin')))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverterErrors(t *testing.T) {
	columnMap := map[string]string{
		"age": "users.age",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name        string
		celExpr     string
		expectError bool
	}{
		{
			name:        "unknown column",
			celExpr:     "unknown_field == 5",
			expectError: true,
		},
		{
			name:        "invalid syntax - incomplete expression",
			celExpr:     "age == ",
			expectError: true,
		},
		{
			name:        "invalid syntax - missing operand",
			celExpr:     "&& age > 18",
			expectError: true,
		},
		{
			name:        "empty expression",
			celExpr:     "",
			expectError: true,
		},
		{
			name:        "invalid operator",
			celExpr:     "age === 18",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			if tt.expectError {
				assert.Error(t, err, "Expected error for expression: %s", tt.celExpr)
				assert.Empty(t, result, "Result should be empty on error")
			} else {
				assert.NoError(t, err, "Should not error for expression: %s", tt.celExpr)
			}
		})
	}
}

func TestConverterUnknownColumns(t *testing.T) {
	columnMap := map[string]string{
		"age":  "users.age",
		"name": "users.name",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	t.Run("reference unknown column", func(t *testing.T) {
		_, err := conv.Convert("email == 'test@example.com'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undeclared reference")
	})

	t.Run("known and unknown columns", func(t *testing.T) {
		_, err := conv.Convert("age > 18 && unknown_col == 'value'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undeclared reference")
	})
}

func TestConverterEdgeCases(t *testing.T) {
	columnMap := map[string]string{
		"value":  "table.value",
		"status": "table.status",
	}

	conv, err := cel2sqlite.NewConverterSimple(columnMap)
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "empty string comparison",
			celExpr:  "status == ''",
			expected: "(table.status = '')",
		},
		{
			name:     "zero comparison",
			celExpr:  "value == 0",
			expected: "(table.value = 0)",
		},
		{
			name:     "parentheses preservation",
			celExpr:  "(value > 10)",
			expected: "(table.value > 10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewConverterErrors(t *testing.T) {
	t.Run("empty column map", func(t *testing.T) {
		conv, err := cel2sqlite.NewConverterSimple(map[string]string{})
		assert.NoError(t, err, "Empty column map should not error")
		assert.NotNil(t, conv)
	})

	t.Run("nil column map", func(t *testing.T) {
		conv, err := cel2sqlite.NewConverterSimple(nil)
		assert.NoError(t, err, "Nil column map should not error")
		assert.NotNil(t, conv)
	})
}

func TestJson(t *testing.T) {
	conv, err := cel2sqlite.NewConverter(map[string]cel2sqlite.ColumnInfo{
		"foo": {SQLName: "foo", Type: cel2sqlite.ColumnTypeJSON},
	}, []string{})
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "empty string comparison",
			celExpr:  "foo == ''",
			expected: "(foo = '')",
		},
		{
			name:     "empty string comparison",
			celExpr:  "foo.ru != ''",
			expected: "(json_extract(foo, '$.ru') != '')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOverridable(t *testing.T) {
	conv, err := cel2sqlite.NewConverter(map[string]cel2sqlite.ColumnInfo{
		"foo": {SQLName: "foo", Type: cel2sqlite.ColumnTypeOverridableField},
	}, []string{"amd64_centos|amd64|"})
	require.NoError(t, err, "Failed to create converter")

	tests := []struct {
		name     string
		celExpr  string
		expected string
	}{
		{
			name:     "empty string comparison",
			celExpr:  "foo == ''",
			expected: "(overridable_resolve(foo, 'amd64_centos|amd64|') = '')",
		},
		{
			name:     "empty string comparison",
			celExpr:  "foo.raw == ''",
			expected: "(foo = '')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.celExpr)
			require.NoError(t, err, "Convert should not fail")
			assert.Equal(t, tt.expected, result)
		})
	}
}
