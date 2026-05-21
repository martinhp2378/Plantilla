package pagination

import (
	"net/url"
	"strings"
	"testing"
)

func TestNewCriteriosPaginaDefault(t *testing.T) {
	c := NewCriteriosPaginaDefault[int]()

	if c.Index != 1 {
		t.Fatalf("Index = %d, want 1", c.Index)
	}
	if c.Size != 10 {
		t.Fatalf("Size = %d, want 10", c.Size)
	}
	if c.First != 0 {
		t.Fatalf("First = %d, want 0", c.First)
	}
	if c.Direction != "asc" {
		t.Fatalf("Direction = %q, want asc", c.Direction)
	}
	if c.FilteredBy == nil {
		t.Fatal("FilteredBy is nil, want initialized map")
	}
}

func TestWithSort(t *testing.T) {
	t.Run("set new column starts in asc", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithSort("nombre")

		if c.SortedBy != "nombre" {
			t.Fatalf("SortedBy = %q, want nombre", c.SortedBy)
		}
		if c.Direction != "asc" {
			t.Fatalf("Direction = %q, want asc", c.Direction)
		}
	})

	t.Run("same column toggles direction", func(t *testing.T) {
		base := CriteriosPagina[int]{SortedBy: "nombre", Direction: "asc"}

		toggled := base.WithSort("nombre")
		if toggled.Direction != "desc" {
			t.Fatalf("Direction after first toggle = %q, want desc", toggled.Direction)
		}

		toggled = toggled.WithSort("nombre")
		if toggled.Direction != "asc" {
			t.Fatalf("Direction after second toggle = %q, want asc", toggled.Direction)
		}
	})
}

func TestWithFilter(t *testing.T) {
	t.Run("adds and updates non empty values", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithFilter("region", "toluca")
		c = c.WithFilter("activa", true)

		if got := c.FilteredBy["region"]; got != "toluca" {
			t.Fatalf("filter region = %v, want toluca", got)
		}
		if got := c.FilteredBy["activa"]; got != true {
			t.Fatalf("filter activa = %v, want true", got)
		}
	})

	t.Run("empty string removes key", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithFilter("region", "toluca")
		c = c.WithFilter("region", "")

		if _, ok := c.FilteredBy["region"]; ok {
			t.Fatal("region filter still present, want removed")
		}
	})

	t.Run("nil map is initialized", func(t *testing.T) {
		c := CriteriosPagina[int]{}
		c = c.WithFilter("x", 1)

		if c.FilteredBy == nil {
			t.Fatal("FilteredBy is nil, want initialized map")
		}
		if got := c.FilteredBy["x"]; got != 1 {
			t.Fatalf("filter x = %v, want 1", got)
		}
	})
}

func TestWithPageAndNavigation(t *testing.T) {
	t.Run("with page calculates offset", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithPage(3)

		if c.Index != 3 {
			t.Fatalf("Index = %d, want 3", c.Index)
		}
		if c.First != 20 {
			t.Fatalf("First = %d, want 20", c.First)
		}
	})

	t.Run("page lower than 1 is clamped", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithPage(0)

		if c.Index != 1 || c.First != 0 {
			t.Fatalf("got Index=%d First=%d, want Index=1 First=0", c.Index, c.First)
		}
	})

	t.Run("next and previous page", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithPage(2)

		next := c.NextPage()
		if next.Index != 3 || next.First != 20 {
			t.Fatalf("next got Index=%d First=%d, want Index=3 First=20", next.Index, next.First)
		}

		prev := c.PreviousPage()
		if prev.Index != 1 || prev.First != 0 {
			t.Fatalf("previous got Index=%d First=%d, want Index=1 First=0", prev.Index, prev.First)
		}
	})

	t.Run("compatibility helpers delegate to navigation", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]().WithPage(2)

		forward := c.WithPageFoward(99)
		if forward.Index != 3 {
			t.Fatalf("WithPageFoward index = %d, want 3", forward.Index)
		}

		backward := c.WithPageBackward(99)
		if backward.Index != 1 {
			t.Fatalf("WithPageBackward index = %d, want 1", backward.Index)
		}
	})
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		name  string
		count int
		size  int
		want  int
	}{
		{name: "exact division", count: 20, size: 10, want: 2},
		{name: "ceil division", count: 21, size: 10, want: 3},
		{name: "zero size returns zero", count: 21, size: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CriteriosPagina[int]{Count: tt.count, Size: tt.size}
			if got := c.TotalPages(); got != tt.want {
				t.Fatalf("TotalPages() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSortIcon(t *testing.T) {
	c := CriteriosPagina[int]{SortedBy: "nombre", Direction: "asc"}

	if got := c.SortIcon("otra"); got != "↕" {
		t.Fatalf("SortIcon for unsorted column = %q, want ↕", got)
	}
	if got := c.SortIcon("nombre"); got != "▲" {
		t.Fatalf("SortIcon asc = %q, want ▲", got)
	}

	c.Direction = "DESC"
	if got := c.SortIcon("nombre"); got != "▼" {
		t.Fatalf("SortIcon desc = %q, want ▼", got)
	}
}

func TestBuildURL(t *testing.T) {
	t.Run("preserves paging sorting and filters", func(t *testing.T) {
		c := CriteriosPagina[int]{
			Size:      25,
			SortedBy:  "nombre",
			Direction: "DESC",
			FilteredBy: map[string]any{
				"region": "toluca",
				"activo": true,
				"omit":   nil,
			},
		}

		url := c.BuildURL("/encuestas", 4)

		assertContainsAll(t, url,
			"/encuestas?",
			"page=4",
			"size=25",
			"sort=nombre",
			"dir=desc",
			"region=toluca",
			"activo=true",
		)
		if strings.Contains(url, "omit=") {
			t.Fatalf("URL contains nil filter value: %s", url)
		}
	})

	t.Run("uses ampersand when base has query", func(t *testing.T) {
		c := NewCriteriosPaginaDefault[int]()
		url := c.BuildURL("/encuestas?tab=abiertas", 2)

		if !strings.Contains(url, "?tab=abiertas&page=2") && !strings.Contains(url, "?page=2&") {
			t.Fatalf("URL should append query params with &, got: %s", url)
		}
		if !strings.Contains(url, "&size=10") {
			t.Fatalf("URL missing size param, got: %s", url)
		}
	})
}

func assertContainsAll(t *testing.T, value string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(value, sub) {
			t.Fatalf("value %q does not contain %q", value, sub)
		}
	}
}

func TestFromQuery(t *testing.T) {
	t.Run("extracts pagination and sort params", func(t *testing.T) {
		q := url.Values{}
		q.Set("page", "3")
		q.Set("size", "25")
		q.Set("sort", "nombre")
		q.Set("dir", "DESC")

		c := FromQuery[int](q)

		if c.Index != 3 {
			t.Fatalf("Index = %d, want 3", c.Index)
		}
		if c.Size != 25 {
			t.Fatalf("Size = %d, want 25", c.Size)
		}
		if c.First != 50 {
			t.Fatalf("First = %d, want 50", c.First)
		}
		if c.SortedBy != "nombre" {
			t.Fatalf("SortedBy = %q, want nombre", c.SortedBy)
		}
		if c.Direction != "desc" {
			t.Fatalf("Direction = %q, want desc", c.Direction)
		}
	})

	t.Run("converts non-reserved params to filters", func(t *testing.T) {
		q := url.Values{}
		q.Set("page", "1")
		q.Set("size", "10")
		q.Set("folio", "F-001")
		q.Set("periodo", "ENERO-JUNIO")
		q.Set("anio", "2025")

		c := FromQuery[int](q)

		if got := c.FilteredBy["folio"]; got != "F-001" {
			t.Fatalf("filter folio = %v, want F-001", got)
		}
		if got := c.FilteredBy["periodo"]; got != "ENERO-JUNIO" {
			t.Fatalf("filter periodo = %v, want ENERO-JUNIO", got)
		}
		if got := c.FilteredBy["anio"]; got != "2025" {
			t.Fatalf("filter anio = %v, want 2025", got)
		}
		// reserved params should NOT be in filters
		if _, ok := c.FilteredBy["page"]; ok {
			t.Fatal("page should not be a filter")
		}
		if _, ok := c.FilteredBy["size"]; ok {
			t.Fatal("size should not be a filter")
		}
		if _, ok := c.FilteredBy["sort"]; ok {
			t.Fatal("sort should not be a filter")
		}
		if _, ok := c.FilteredBy["dir"]; ok {
			t.Fatal("dir should not be a filter")
		}
	})

	t.Run("skips empty filter values", func(t *testing.T) {
		q := url.Values{}
		q.Set("page", "1")
		q.Set("folio", "F-001")
		q.Set("periodo", "") // empty

		c := FromQuery[int](q)

		if _, ok := c.FilteredBy["periodo"]; ok {
			t.Fatal("empty periodo should not be a filter")
		}
		if got := c.FilteredBy["folio"]; got != "F-001" {
			t.Fatalf("filter folio = %v, want F-001", got)
		}
	})

	t.Run("defaults when params are missing", func(t *testing.T) {
		q := url.Values{}

		c := FromQuery[int](q)

		if c.Index != 1 {
			t.Fatalf("Index = %d, want 1", c.Index)
		}
		if c.Size != 10 {
			t.Fatalf("Size = %d, want 10", c.Size)
		}
		if c.First != 0 {
			t.Fatalf("First = %d, want 0", c.First)
		}
		if c.SortedBy != "" {
			t.Fatalf("SortedBy = %q, want empty", c.SortedBy)
		}
		if c.Direction != "" {
			t.Fatalf("Direction = %q, want empty", c.Direction)
		}
	})
}
