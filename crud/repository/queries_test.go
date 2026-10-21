package repository_test

import (
	"PlantillaInstitucional/crud/repository"
	"strings"
	"testing"
)

func TestCrudQueriesStructure(t *testing.T) {
	t.Run("SelectCrud uses alias and no schema prefix", func(t *testing.T) {
		q := repository.SelectCrud
		if !strings.Contains(q, "FROM crud AS c") {
			t.Errorf("expected 'FROM crud AS c' without schema prefix, got: %s", q)
		}
		if !strings.Contains(q, "c.fecha::text AS fecha") {
			t.Errorf("expected date cast 'c.fecha::text AS fecha', got: %s", q)
		}
		if !strings.Contains(q, "c.hora::text AS hora") {
			t.Errorf("expected time cast 'c.hora::text AS hora', got: %s", q)
		}
	})

	t.Run("InsertCrud uses RETURNING id", func(t *testing.T) {
		q := repository.InsertCrud
		if !strings.Contains(q, "INSERT INTO crud") {
			t.Errorf("expected 'INSERT INTO crud', got: %s", q)
		}
		if !strings.Contains(q, "RETURNING id") {
			t.Errorf("expected 'RETURNING id', got: %s", q)
		}
	})

	t.Run("UpdateCrud targets id parameter", func(t *testing.T) {
		q := repository.UpdateCrud
		if !strings.Contains(q, "UPDATE crud SET") {
			t.Errorf("expected 'UPDATE crud SET', got: %s", q)
		}
		if !strings.Contains(q, "WHERE id = $15") {
			t.Errorf("expected 'WHERE id = $15', got: %s", q)
		}
	})
}
