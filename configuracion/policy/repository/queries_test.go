package repository

import (
	"strings"
	"testing"
)

func TestMatchPathPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		path   string
		want   bool
	}{
		{
			name:   "exact match",
			prefix: "/dashboard",
			path:   "/dashboard",
			want:   true,
		},
		{
			name:   "sub-path match",
			prefix: "/dashboard",
			path:   "/dashboard/reports",
			want:   true,
		},
		{
			name:   "partial prefix string match bypass attempt",
			prefix: "/dashboard",
			path:   "/dashboard-secreto",
			want:   false,
		},
		{
			name:   "nested sub-path match",
			prefix: "/dashboard",
			path:   "/dashboard/reports/weekly/summary",
			want:   true,
		},
		{
			name:   "completely different path",
			prefix: "/dashboard",
			path:   "/home",
			want:   false,
		},
		{
			name:   "prefix is substring of path but not segment matched",
			prefix: "/a",
			path:   "/ab",
			want:   false,
		},
		{
			name:   "root prefix matches everything",
			prefix: "/",
			path:   "/dashboard",
			want:   true,
		},
		{
			name:   "empty prefix does not match",
			prefix: "",
			path:   "/dashboard",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPathPrefix(tt.prefix, tt.path)
			if got != tt.want {
				t.Errorf("MatchPathPrefix(%q, %q) = %v; want %v", tt.prefix, tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchRoutePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "literal match without parameters",
			pattern: "/institucion/personas",
			path:    "/institucion/personas",
			want:    true,
		},
		{
			name:    "single parameter match",
			pattern: "/institucion/personas/{id}",
			path:    "/institucion/personas/123",
			want:    true,
		},
		{
			name:    "parameter does not cross segment boundary",
			pattern: "/institucion/personas/{id}",
			path:    "/institucion/personas/123/extra",
			want:    false,
		},
		{
			name:    "nested multiple parameters match",
			pattern: "/departamentos/{dep_id}/personas/{id}",
			path:    "/departamentos/42/personas/abc-uuid",
			want:    true,
		},
		{
			name:    "different static segment fails",
			pattern: "/institucion/personas/{id}",
			path:    "/institucion/usuarios/123",
			want:    false,
		},
		{
			name:    "trailing slashes tolerated",
			pattern: "/institucion/personas/{id}/",
			path:    "/institucion/personas/123",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchRoutePattern(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("MatchRoutePattern(%q, %q) = %v; want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestQueriesIncludeReglasAndNavigation(t *testing.T) {
	if !strings.Contains(listPoliticasBase, "p.reglas::text AS reglas") {
		t.Error("listPoliticasBase must select reglas")
	}
	if !strings.Contains(routeAllowedQuery, "jsonb_array_elements") {
		t.Error("routeAllowedQuery must parse routes jsonb array")
	}
	if !strings.Contains(navigationPathsQuery, "route->>'path'") {
		t.Error("navigationPathsQuery must project literal routes")
	}
}
