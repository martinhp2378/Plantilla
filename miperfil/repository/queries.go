package repository

const SelectPerfilUsuario = `
	SELECT 
		p.uuid_persona,
		p.nombre,
		p.apellido1,
		p.apellido2,
		p.clave_issemym,
		p.curp,
		p.rfc,
		COALESCE(p.imagen_perfil, 'placeholder.png') AS imagen_perfil,
		u.email,
		u.perfil
	FROM institucion.persona p
	LEFT JOIN sesion.usuario u ON p.uuid_persona = u.uuid_persona
	WHERE p.uuid_persona = $1
`

const SelectPasswordHash = `
	SELECT password FROM sesion.usuario WHERE uuid_persona = $1
`

const UpdatePasswordUsuario = `
	UPDATE sesion.usuario SET
		password = $1,
		razon_m = 'Actualización de password por usuario',
		fecha_m = now()
	WHERE uuid_persona = $2
`

const UpdateImagenPersona = `
	UPDATE institucion.persona SET
		imagen_perfil = $1,
		razon_m = 'Actualización de avatar por usuario',
		fecha_m = now()
	WHERE uuid_persona = $2
`
