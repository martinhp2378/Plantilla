package repository

const routeAllowedQuery = `
SELECT EXISTS (
    SELECT 1
    FROM sesion.politica p
    WHERE p.perfil = $1
      AND $2 LIKE (p.reglas->>'path_prefix') || '%'
)`

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
    COALESCE(p.reglas->>'path_prefix', '') AS path_prefix
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
VALUES ($1, $2, $3, $4, $5, jsonb_build_object('path_prefix', $6))
RETURNING nombre
`

const updatePolitica = `
UPDATE sesion.politica
SET nombre = $1,
    descripcion = $2,
    perfil = $3,
    recurso = $4,
    accion = $5,
    reglas = jsonb_build_object('path_prefix', $6)
WHERE nombre = $7
`

const deletePolitica = `
DELETE FROM sesion.politica
WHERE nombre = $1
`

const listPerfiles = `SELECT nombre FROM sesion.perfil ORDER BY nombre`
const listRecursos = `SELECT nombre FROM sesion.recurso ORDER BY nombre`
const listAcciones = `SELECT nombre FROM sesion.cat_accion ORDER BY nombre`
