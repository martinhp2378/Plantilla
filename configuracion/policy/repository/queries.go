package repository

import "strings"

// Literal routes accept one optional final slash, not path-prefix expansion.
// Keep method/profile/action checks outside that equivalence. Parameter patterns
// use PostgreSQL regex escaping (this is a Go raw string, not a quoted string).
const routeAllowedQuery = `
SELECT EXISTS (
    SELECT 1
    FROM sesion.politica p
    CROSS JOIN LATERAL jsonb_array_elements(COALESCE(p.reglas->'routes', '[]'::jsonb)) route
    CROSS JOIN LATERAL (
        SELECT
            string_to_array(trim(both '/' from route->>'pattern'), '/') AS pattern_segments,
            string_to_array(trim(both '/' from $3), '/') AS path_segments
    ) segments
    WHERE p.perfil = $1
      AND (route->>'method' = '*' OR upper(route->>'method') = upper($2))
      AND (COALESCE(route->>'accion', '') = '' OR route->>'accion' = p.accion)
      AND (
          route->>'path' = $3 OR
          (
              route->>'path' LIKE '/%' AND $3 LIKE '/%' AND
              route->>'path' NOT LIKE '%//%' AND $3 NOT LIKE '%//%' AND
              regexp_replace(route->>'path', '/$', '') = regexp_replace($3, '/$', '')
          ) OR
		  (
		      COALESCE(route->>'prefix', '') <> '' AND
		      (route->>'prefix' = '/' OR $3 = route->>'prefix' OR starts_with($3, (route->>'prefix') || '/'))
		  ) OR
          (
              COALESCE(route->>'pattern', '') <> '' AND
              array_length(segments.pattern_segments, 1) = array_length(segments.path_segments, 1) AND
              NOT EXISTS (
                  SELECT 1
                  FROM generate_subscripts(segments.pattern_segments, 1) i
                  WHERE segments.pattern_segments[i] !~ '^\{[^/]+\}$'
                    AND segments.pattern_segments[i] <> segments.path_segments[i]
              )
          )
      )
    UNION ALL
    SELECT 1
    FROM sesion.politica p
    WHERE p.perfil = $1
      AND NOT (p.reglas ? 'routes')
      AND COALESCE(p.reglas->>'path_prefix', '') <> ''
      AND (
          p.reglas->>'path_prefix' = '/' 
          OR $3 = p.reglas->>'path_prefix' 
          OR starts_with($3, (p.reglas->>'path_prefix') || '/')
      )
)`

func MatchPathPrefix(prefix, path string) bool {
	if prefix == "" {
		return false
	}
	if prefix == "/" {
		return strings.HasPrefix(path, "/")
	}
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, prefix+"/")
}

func MatchRoutePattern(pattern, path string) bool {
	pattern = strings.Trim(pattern, "/")
	path = strings.Trim(path, "/")
	if pattern == "" || path == "" {
		return pattern == path
	}

	patternSegments := strings.Split(pattern, "/")
	pathSegments := strings.Split(path, "/")
	if len(patternSegments) != len(pathSegments) {
		return false
	}

	for i := range patternSegments {
		segment := patternSegments[i]
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(segment) > 2 {
			continue
		}
		if segment != pathSegments[i] {
			return false
		}
	}

	return true
}

// Only explicit literal GET consultar routes can become navigation grants.
const navigationPathsQuery = `
SELECT DISTINCT route->>'path'
FROM sesion.politica p
CROSS JOIN LATERAL jsonb_array_elements(COALESCE(p.reglas->'routes', '[]'::jsonb)) route
WHERE p.perfil = $1
  AND p.accion = 'consultar'
  AND (route->>'method' = '*' OR upper(route->>'method') = 'GET')
  AND (COALESCE(route->>'accion', '') = '' OR route->>'accion' = p.accion)
  AND route->>'path' LIKE '/%'
  AND route->>'path' !~ '[{}*?:#%]'
ORDER BY 1
`

const profileHasPoliciesQuery = `
SELECT EXISTS (
    SELECT 1
    FROM sesion.perfil pe
    WHERE pe.nombre = $1
      AND EXISTS (
        SELECT 1 FROM sesion.politica po WHERE po.perfil = pe.nombre
      )
)`

const listPoliticasBase = `
SELECT
    p.nombre,
    p.descripcion,
    p.perfil,
    p.recurso,
    p.accion,
    COALESCE(p.reglas->>'path_prefix', '') AS path_prefix,
    p.reglas::text AS reglas
FROM sesion.politica p
`

const listPoliticasFiltered = listPoliticasBase + `
%s
ORDER BY %s
LIMIT $%d OFFSET $%d
`

const countPoliticas = `
SELECT COUNT(*)
FROM sesion.politica p
%s
`

const getPoliticaByNombre = listPoliticasBase + `
WHERE p.nombre = $1
`

const insertPolitica = `
INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas)
VALUES (
    $1, $2, $3, $4, $5,
    CASE 
        WHEN $6::text <> '' AND $6::text IS NOT NULL AND jsonb_typeof($6::jsonb) = 'object' THEN $6::jsonb
        ELSE jsonb_build_object('path_prefix', $7::text, 'routes', jsonb_build_array(jsonb_build_object('method', '*', 'prefix', $7::text, 'accion', $5::text)))
    END
)
RETURNING nombre
`

const updatePolitica = `
UPDATE sesion.politica
SET nombre = $1,
    descripcion = $2,
    perfil = $3,
    recurso = $4,
    accion = $5,
    reglas = CASE 
        WHEN $6::text <> '' AND $6::text IS NOT NULL AND jsonb_typeof($6::jsonb) = 'object' THEN $6::jsonb
        ELSE jsonb_build_object('path_prefix', $7::text, 'routes', jsonb_build_array(jsonb_build_object('method', '*', 'prefix', $7::text, 'accion', $5::text)))
    END,
    fecha_m = NOW()
WHERE nombre = $8
`

const deletePolitica = `
DELETE FROM sesion.politica
WHERE nombre = $1
`

const listPerfiles = `SELECT nombre FROM sesion.perfil ORDER BY nombre`
const listRecursos = `SELECT nombre FROM sesion.recurso ORDER BY nombre`
const listAcciones = `SELECT nombre FROM sesion.cat_accion ORDER BY nombre`
