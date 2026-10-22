package repository

const (
	SelectElverFiltered = `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m, COUNT(*) OVER() AS total_count
		FROM sistema.elver
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`

	SelectElverByID = `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m
		FROM sistema.elver
		WHERE id = $1
	`

	SelectAllElver = `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m
		FROM sistema.elver
		ORDER BY fecha_i DESC
	`

	InsertElver = `
		INSERT INTO sistema.elver (nombre, codigo, descripcion, is_active, tenant_id, razon_i)
		VALUES ($1, $2, $3, $4, CASE WHEN current_setting('app.tenant_id', true) ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN current_setting('app.tenant_id', true)::uuid ELSE NULL END, 'Creacion')
		RETURNING id
	`

	UpdateElver = `
		UPDATE sistema.elver
		SET nombre = $1, codigo = $2, descripcion = $3, is_active = $4, razon_m = 'Actualizacion'
		WHERE id = $5
	`

	DeleteElver = `
		DELETE FROM sistema.elver
		WHERE id = $1
	`
)
