package db

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestBuildDynamicWhere_TextFilterWithAlias(t *testing.T) {
	filters := map[string]any{"nombre": "ana"}
	config := FilterConfig{
		Texto: map[string]bool{"nombre": true},
	}

	conditions, args := BuildDynamicWhere(filters, config, "u")

	if len(conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(conditions))
	}
	if conditions[0] != "u.nombre ILIKE $1" {
		t.Fatalf("unexpected condition: %q", conditions[0])
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != "%ana%" {
		t.Fatalf("unexpected arg: %#v", args[0])
	}
}

func TestBuildDynamicWhere_NumericFilterParsesString(t *testing.T) {
	filters := map[string]any{"edad": "25"}
	config := FilterConfig{
		Numericas: map[string]bool{"edad": true},
	}

	conditions, args := BuildDynamicWhere(filters, config, "")

	if len(conditions) != 1 || conditions[0] != "edad = $1" {
		t.Fatalf("unexpected conditions: %#v", conditions)
	}
	if len(args) != 1 || args[0] != 25 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildDynamicWhere_UUIDFilterFromString(t *testing.T) {
	rawUUID := "550e8400-e29b-41d4-a716-446655440000"
	parsedUUID, err := uuid.FromString(rawUUID)
	if err != nil {
		t.Fatalf("setup uuid failed: %v", err)
	}

	filters := map[string]any{"usuario_id": rawUUID}
	config := FilterConfig{
		UUID: map[string]bool{"usuario_id": true},
	}

	conditions, args := BuildDynamicWhere(filters, config, "u")

	if len(conditions) != 1 || conditions[0] != "u.usuario_id = $1" {
		t.Fatalf("unexpected conditions: %#v", conditions)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if argUUID, ok := args[0].(uuid.UUID); !ok || argUUID != parsedUUID {
		t.Fatalf("unexpected uuid arg: %#v", args[0])
	}
}

func TestBuildDynamicWhere_DateRangeVariants(t *testing.T) {
	start := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		value              any
		expectedCondition  string
		expectedArgsLength int
	}{
		{
			name:               "between when start and end are present",
			value:              DateRange{Start: start, End: end},
			expectedCondition:  "fecha BETWEEN $1 AND $2",
			expectedArgsLength: 2,
		},
		{
			name:               "greater than or equal when only start is present",
			value:              DateRange{Start: start},
			expectedCondition:  "fecha >= $1",
			expectedArgsLength: 1,
		},
		{
			name:               "less than or equal when only end is present",
			value:              DateRange{End: end},
			expectedCondition:  "fecha <= $1",
			expectedArgsLength: 1,
		},
		{
			name:               "exact equality for time value",
			value:              start,
			expectedCondition:  "fecha = $1",
			expectedArgsLength: 1,
		},
	}

	config := FilterConfig{Fecha: map[string]bool{"fecha": true}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filters := map[string]any{"fecha": tc.value}

			conditions, args := BuildDynamicWhere(filters, config, "")

			if len(conditions) != 1 || conditions[0] != tc.expectedCondition {
				t.Fatalf("unexpected conditions: %#v", conditions)
			}
			if len(args) != tc.expectedArgsLength {
				t.Fatalf("unexpected args length: %d", len(args))
			}
		})
	}
}

func TestBuildDynamicWhere_BooleanFilterFromInt(t *testing.T) {
	filters := map[string]any{"activo": 1}
	config := FilterConfig{Boolean: map[string]bool{"activo": true}}

	conditions, args := BuildDynamicWhere(filters, config, "")

	if len(conditions) != 1 || conditions[0] != "activo = $1" {
		t.Fatalf("unexpected conditions: %#v", conditions)
	}
	if len(args) != 1 || args[0] != true {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildDynamicWhere_SkipsNilAndUnsupportedValues(t *testing.T) {
	filters := map[string]any{
		"nombre":     nil,
		"edad":       "0",
		"usuario_id": "invalid-uuid",
	}
	config := FilterConfig{
		Texto:     map[string]bool{"nombre": true},
		Numericas: map[string]bool{"edad": true},
		UUID:      map[string]bool{"usuario_id": true},
	}

	conditions, args := BuildDynamicWhere(filters, config, "")

	if len(conditions) != 0 {
		t.Fatalf("expected no conditions, got %#v", conditions)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %#v", args)
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected int
		ok       bool
	}{
		{name: "int", value: 7, expected: 7, ok: true},
		{name: "float64", value: 9.9, expected: 9, ok: true},
		{name: "numeric string", value: "42", expected: 42, ok: true},
		{name: "empty string", value: "", expected: 0, ok: false},
		{name: "invalid string", value: "abc", expected: 0, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ToInt(tc.value)
			if got != tc.expected || ok != tc.ok {
				t.Fatalf("got (%d,%v), expected (%d,%v)", got, ok, tc.expected, tc.ok)
			}
		})
	}
}

func TestToUUID(t *testing.T) {
	rawUUID := "550e8400-e29b-41d4-a716-446655440000"
	parsedUUID, err := uuid.FromString(rawUUID)
	if err != nil {
		t.Fatalf("setup uuid failed: %v", err)
	}

	tests := []struct {
		name     string
		value    any
		expected uuid.UUID
		ok       bool
	}{
		{name: "uuid type", value: parsedUUID, expected: parsedUUID, ok: true},
		{name: "valid string", value: rawUUID, expected: parsedUUID, ok: true},
		{name: "empty string", value: "", expected: uuid.Nil, ok: false},
		{name: "invalid string", value: "abc", expected: uuid.Nil, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ToUUID(tc.value)
			if got != tc.expected || ok != tc.ok {
				t.Fatalf("got (%v,%v), expected (%v,%v)", got, ok, tc.expected, tc.ok)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
		ok       bool
	}{
		{name: "bool true", value: true, expected: true, ok: true},
		{name: "string true", value: "true", expected: true, ok: true},
		{name: "int one", value: 1, expected: true, ok: true},
		{name: "int zero", value: 0, expected: false, ok: true},
		{name: "invalid int", value: 2, expected: false, ok: false},
		{name: "invalid string", value: "yes", expected: false, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ToBool(tc.value)
			if got != tc.expected || ok != tc.ok {
				t.Fatalf("got (%v,%v), expected (%v,%v)", got, ok, tc.expected, tc.ok)
			}
		})
	}
}
