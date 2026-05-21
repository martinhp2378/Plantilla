package convertidores

import (
	"reflect"
	"testing"

	"PlantillaInstitucional/shared/views/components"
)

func TestOpcionesDeEstadoYBooleanos(t *testing.T) {
	if got, want := OptionFromEstatusGenerico(), []components.SelectOption{{Value: "", Text: "Seleccione un Estatus"}, {Value: "1", Text: "Activo"}, {Value: "-1", Text: "Inactivo"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("OptionFromEstatusGenerico() = %#v, want %#v", got, want)
	}

	if got, want := OptionFromBooleanoGenerico(), []components.SelectOption{{Value: "", Text: "Seleccione una Opcion"}, {Value: "1", Text: "Si"}, {Value: "-1", Text: "No"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("OptionFromBooleanoGenerico() = %#v, want %#v", got, want)
	}

	tests := []struct {
		name  string
		input bool
		want  string
	}{
		{name: "true", input: true, want: "1"},
		{name: "false", input: false, want: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Traducir(tt.input); got != tt.want {
				t.Fatalf("Traducir(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
