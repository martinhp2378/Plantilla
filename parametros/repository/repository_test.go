package repository

import (
	shared "PlantillaInstitucional/shared/db"
	"testing"

	"github.com/gofrs/uuid/v5"
)




func TestBuildOrderByClause(t *testing.T) {
	t.Run("default order", func(t *testing.T) {
		got, err := shared.BuildOrderBy("", "", colsOrden, "parametro", "clave ASC")
		if err != nil {
			t.Fatalf("BuildOrderBy() unexpected error: %v", err)
		}
		if got != "clave ASC" {
			t.Fatalf("BuildOrderBy() = %q, want %q", got, "clave ASC")
		}
	})

	t.Run("valid explicit order", func(t *testing.T) {
		got, err := shared.BuildOrderBy("clave", "DESC", colsOrden, "parametro", "clave ASC")
		if err != nil {
			t.Fatalf("BuildOrderBy() unexpected error: %v", err)
		}
		if got != "parametro.clave DESC" {
			t.Fatalf("BuildOrderBy() = %q, want %q", got, "parametro.clave DESC")
		}
	})

	t.Run("invalid column", func(t *testing.T) {
		_, err := shared.BuildOrderBy("DROP_TABLE", "", colsOrden, "parametro", "clave ASC")
		if err == nil {
			t.Fatal("BuildOrderBy() expected error for invalid column")
		}
	})
}

func TestBuildDynamicWhere(t *testing.T) {
	config := shared.FilterConfig{
		Texto:     colsTexto,
		Numericas: colsNumericas,
		UUID:      colsUUID,
	}

	t.Run("text filter", func(t *testing.T) {
		condiciones, args := shared.BuildDynamicWhere(map[string]any{"clave": "TOKEN"}, config, "parametro")
		if len(condiciones) != 1 {
			t.Fatalf("len(condiciones) = %d, want %d", len(condiciones), 1)
		}
		if condiciones[0] != "parametro.clave ILIKE $1" {
			t.Fatalf("condicion = %q, want %q", condiciones[0], "parametro.clave ILIKE $1")
		}
		if len(args) != 1 || args[0] != "%TOKEN%" {
			t.Fatalf("args = %#v, want [%q]", args, "%TOKEN%")
		}
	})

	t.Run("uuid filter", func(t *testing.T) {
		id := uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440000"))
		condiciones, args := shared.BuildDynamicWhere(map[string]any{"parametro_id": id.String()}, config, "parametro")
		if len(condiciones) != 1 {
			t.Fatalf("len(condiciones) = %d, want %d", len(condiciones), 1)
		}
		if condiciones[0] != "parametro.parametro_id = $1" {
			t.Fatalf("condicion = %q, want %q", condiciones[0], "parametro.parametro_id = $1")
		}
		if len(args) != 1 {
			t.Fatalf("len(args) = %d, want %d", len(args), 1)
		}
		parsed, ok := args[0].(uuid.UUID)
		if !ok || parsed != id {
			t.Fatalf("args[0] = %#v, want uuid %s", args[0], id.String())
		}
	})

	t.Run("skip nil empty and unknown", func(t *testing.T) {
		condiciones, args := shared.BuildDynamicWhere(map[string]any{
			"clave":        "",
			"unknown":      "abc",
			"parametro_id": nil,
		}, config, "parametro")
		if len(condiciones) != 0 {
			t.Fatalf("len(condiciones) = %d, want %d", len(condiciones), 0)
		}
		if len(args) != 0 {
			t.Fatalf("len(args) = %d, want %d", len(args), 0)
		}
	})
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
		ok   bool
	}{
		{name: "int", in: 7, want: 7, ok: true},
		{name: "float64", in: 7.9, want: 7, ok: true},
		{name: "string", in: "9", want: 9, ok: true},
		{name: "empty string", in: "", want: 0, ok: false},
		{name: "invalid string", in: "abc", want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := shared.ToInt(tt.in)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("ToInt(%#v) = (%d, %v), want (%d, %v)", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestToUUID(t *testing.T) {
	id := uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440000"))

	tests := []struct {
		name string
		in   any
		want uuid.UUID
		ok   bool
	}{
		{name: "uuid", in: id, want: id, ok: true},
		{name: "uuid string", in: id.String(), want: id, ok: true},
		{name: "empty string", in: "", want: uuid.Nil, ok: false},
		{name: "invalid string", in: "not-uuid", want: uuid.Nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := shared.ToUUID(tt.in)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("ToUUID(%#v) = (%s, %v), want (%s, %v)", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}
