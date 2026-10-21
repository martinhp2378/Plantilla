package crud_test

import (
	"PlantillaInstitucional/crud"
	"testing"
)

func TestCrudToCriteria(t *testing.T) {
	c := crud.Crud{
		TextoCorto: "Ejemplo",
		Seleccion:  "opcion_a",
		Correo:     "test@issemym.gob.mx",
	}

	crit := c.ToCriteria(2, 15, "texto_corto", "asc")

	if crit.Index != 2 {
		t.Errorf("expected page 2, got %d", crit.Index)
	}
	if crit.Size != 15 {
		t.Errorf("expected size 15, got %d", crit.Size)
	}
	if crit.First != 15 {
		t.Errorf("expected offset 15, got %d", crit.First)
	}
	if crit.SortedBy != "texto_corto" {
		t.Errorf("expected sort by texto_corto, got %s", crit.SortedBy)
	}
	if crit.Direction != "asc" {
		t.Errorf("expected direction asc, got %s", crit.Direction)
	}
	if crit.FilteredBy["texto_corto"] != "Ejemplo" {
		t.Errorf("expected filter texto_corto 'Ejemplo', got %v", crit.FilteredBy["texto_corto"])
	}
	if crit.FilteredBy["seleccion"] != "opcion_a" {
		t.Errorf("expected filter seleccion 'opcion_a', got %v", crit.FilteredBy["seleccion"])
	}
	if crit.FilteredBy["correo"] != "test@issemym.gob.mx" {
		t.Errorf("expected filter correo 'test@issemym.gob.mx', got %v", crit.FilteredBy["correo"])
	}
}
