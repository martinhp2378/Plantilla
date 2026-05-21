package repository

const SelectParametrosInicial = `
SELECT 
	parametro.clave,
	parametro.valor
FROM cfg_parametros AS parametro
WHERE parametro.is_active = true
`

const SelectParametros = `
SELECT 
	parametro.parametro_id,
	parametro.clave,
	parametro.valor,
	parametro.descripcion,
	parametro.tipo_dato,
	parametro.is_active,
	parametro.razon_i,
	parametro.razon_m,
	parametro.fecha_i,
	parametro.fecha_m
FROM cfg_parametros AS parametro
`

// SelectParametrosFiltered
// %s (1): WHERE clause
// %s (2): ORDER BY column + direction
// %d (3): LIMIT
// %d (4): OFFSET
const SelectParametrosFiltered = SelectParametros + `
%s 
ORDER BY %s 
LIMIT $%d OFFSET $%d
`

const SelectParametroXId = SelectParametros + `
WHERE parametro.parametro_id = $1
`

const SelectParametroValorByClave = `
SELECT parametro.valor
FROM cfg_parametros AS parametro
WHERE parametro.clave = $1
  AND parametro.is_active = true
`

const SelectParametrosIds = `SELECT parametro_id FROM cfg_parametros`

// CountParametros
// %s: WHERE clause
const CountParametros = `
SELECT COUNT(*) 
FROM cfg_parametros AS parametro
%s`

const InsertParametro = `
INSERT INTO cfg_parametros
(parametro_id,
clave, 
valor,
descripcion,
tipo_dato,
is_active,
razon_i) 
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING parametro_id
`

const UpdateParametro = `
UPDATE cfg_parametros SET 
clave = $1, 
valor = $2,
descripcion = $3,
tipo_dato = $4,
is_active = $5,
razon_m = $6,
fecha_m = NOW()
WHERE parametro_id = $7
`

const DeleteParametro = `DELETE FROM cfg_parametros WHERE parametro_id = $1`
