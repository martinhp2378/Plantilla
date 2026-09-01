package repository

// Usamos backticks para strings multilinea limpios.

const SelectUsuarios = `
	SELECT 
		u.uuid_persona,
		u.email,
		u.password,
		u.perfil,
		u.razon_i,
		u.razon_m,
		u.fecha_i,
		u.fecha_m,
		CONCAT_WS(' ', p.nombre, p.apellido1, p.apellido2) AS nombre_completo
	FROM sesion.usuario AS u
	INNER JOIN institucion.persona AS p ON u.uuid_persona = p.uuid_persona
`

// SelectUsuariosFiltered
// %s (1): WHERE clause
// %s (2): ORDER BY column + direction
// %d (3): LIMIT
// %d (4): OFFSET
const SelectUsuariosFiltered = SelectUsuarios + `
	%s 
	ORDER BY %s 
	LIMIT $%d OFFSET $%d
`

const SelectUsuarioXId = SelectUsuarios + `
	WHERE u.uuid_persona = $1
`

// CountUsuarios
// %s: WHERE clause (debe ser idéntico al del Select principal)
const CountUsuarios = `
	SELECT COUNT(*) 
	FROM sesion.usuario AS u
	INNER JOIN institucion.persona AS p ON u.uuid_persona = p.uuid_persona
	%s
`

// InsertUsuario
// Usamos RETURNING id para asegurar que la inserción fue exitosa y obtener el UUID generado (si fuera generado por DB).
const InsertUsuario = `
	INSERT INTO sesion.usuario (
		uuid_persona, email, password, perfil, razon_i, is_active, tenant_id
	) 
	VALUES ($1, $2, $3, $4, $5, false, current_setting('app.tenant_id')::uuid)
	RETURNING uuid_persona
`

const UpdateUsuario = `
	UPDATE sesion.usuario SET
		email = $1, 
		password = $2,
		perfil = $3,
		razon_m = $4
	WHERE uuid_persona = $5
`

const DeleteUsuario = `DELETE FROM sesion.usuario WHERE uuid_persona = $1`

const UpdatePasswordUsuario = `
	UPDATE sesion.usuario SET
		password = $1
	WHERE uuid_persona = $2
`
