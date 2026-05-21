package repository

// Usamos backticks para strings multilinea limpios.

const SelectPersonas = `
	SELECT 
		p.uuid_persona,
		p.nombre,
		p.apellido1,
		p.apellido2,
		p.rfc,
		p.curp,
		p.clave_issemym,
		p.imagen_perfil,
		p.razon_i,
		p.razon_m,
		p.fecha_i,
		p.fecha_m
	FROM persona AS p
`

const SelectPersonasCorto = `
	SELECT 
		p.uuid_persona,
		p.nombre,
		p.apellido1,
		p.apellido2
	FROM persona AS p
	ORDER BY p.nombre, p.apellido1
`

// SelectPersonasFiltered
// %s (1): WHERE clause
// %s (2): ORDER BY column + direction
// %d (3): LIMIT
// %d (4): OFFSET
const SelectPersonasFiltered = SelectPersonas + `
	%s 
	ORDER BY %s 
	LIMIT $%d OFFSET $%d
`

const SelectPersonaXId = SelectPersonas + `
	WHERE p.uuid_persona = $1
`

// CountPersonas
// %s: WHERE clause (debe ser idéntico al del Select principal)
const CountPersonas = `
	SELECT COUNT(*) 
	FROM persona AS p
	%s
`

// InsertPersona
// Usamos RETURNING id para asegurar que la inserción fue exitosa y obtener el UUID generado (si fuera generado por DB).
const InsertPersona = `
	INSERT INTO persona (
		nombre,
		apellido1,
		apellido2,
		rfc,
		curp,
		clave_issemym,
		razon_i
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) 
	RETURNING uuid_persona
`

const UpdatePersona = `
	UPDATE persona SET 
		nombre=$1,
		apellido1=$2,
		apellido2=$3,
		rfc=$4,
		curp=$5,
		clave_issemym = $6,
		razon_m = $7
	WHERE uuid_persona = $8
`

const DeletePersona = `DELETE FROM persona WHERE uuid_persona = $1`
const UpdateImagenPerfil = `
	UPDATE persona SET 
		imagen_perfil=$1
	WHERE uuid_persona = $2
`
