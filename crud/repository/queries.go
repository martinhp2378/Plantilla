package repository

// Consultas SQL base para el repositorio del módulo crud.
// Se evita el prefijo de esquema explícito para permitir la resolución por search_path multi-tenant.

const SelectCrud = `
	SELECT 
		c.id,
		c.texto_corto,
		c.texto_largo,
		c.entero,
		c.decimal,
		c.booleano,
		c.fecha::text AS fecha,
		c.hora::text AS hora,
		c.fecha_hora,
		c.seleccion,
		c.correo,
		c.telefono,
		c.url,
		c.color,
		c.tenant_id,
		c.razon_i,
		c.razon_m,
		c.fecha_i,
		c.fecha_m
	FROM crud AS c
`

const SelectCrudFiltered = SelectCrud + `
	%s 
	ORDER BY %s 
	LIMIT $%d OFFSET $%d
`

const SelectCrudById = SelectCrud + `
	WHERE c.id = $1
`

const CountCrud = `
	SELECT COUNT(*) 
	FROM crud AS c
	%s
`

const InsertCrud = `
	INSERT INTO crud (
		id,
		texto_corto,
		texto_largo,
		entero,
		decimal,
		booleano,
		fecha,
		hora,
		fecha_hora,
		seleccion,
		correo,
		telefono,
		url,
		color,
		razon_i
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7::date, $8::time, $9::timestamptz, $10, $11, $12, $13, $14, $15
	)
	RETURNING id
`

const UpdateCrud = `
	UPDATE crud SET
		texto_corto = $1,
		texto_largo = $2,
		entero = $3,
		decimal = $4,
		booleano = $5,
		fecha = $6::date,
		hora = $7::time,
		fecha_hora = $8::timestamptz,
		seleccion = $9,
		correo = $10,
		telefono = $11,
		url = $12,
		color = $13,
		razon_m = $14
	WHERE id = $15
`

const DeleteCrud = `
	DELETE FROM crud 
	WHERE id = $1
`

const SelectAllCrud = SelectCrud + `
	ORDER BY c.fecha_i DESC
`
