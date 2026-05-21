package repository

// Usamos backticks para strings multilinea limpios.

const SelectPerfiles = `
	SELECT 
		p.nombre
	FROM perfil AS p
`

// SelectPerfilesFiltered
// %s (1): WHERE clause
// %s (2): ORDER BY column + direction
// %d (3): LIMIT
// %d (4): OFFSET
const SelectPerfilesFiltered = SelectPerfiles + `
	%s 
	ORDER BY %s 
	LIMIT $%d OFFSET $%d
`

const SelectPerfilXNombre = SelectPerfiles + `
	WHERE p.nombre = $1
`

// CountPerfiles
// %s: WHERE clause (debe ser idéntico al del Select principal)
const CountPerfiles = `
	SELECT COUNT(*) 
	FROM perfil AS p
	%s
`

// InsertPerfil
// Usamos RETURNING id para asegurar que la inserción fue exitosa y obtener el UUID generado (si fuera generado por DB).
const InsertPerfil = `
	INSERT INTO perfil (
		nombre
	) 
	VALUES ($1) 
	RETURNING nombre
`

const UpdatePerfil = `
	UPDATE perfil SET 
		nombre = $1
	WHERE nombre = $2
`

const DeletePerfil = `DELETE FROM perfil WHERE nombre = $1`
