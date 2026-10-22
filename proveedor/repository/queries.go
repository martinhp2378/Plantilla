package repository

const (
	SelectProveedorFiltered = `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m, COUNT(*) OVER() AS total_count
		FROM sistema.proveedor
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`

	SelectProveedorByID = `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m
		FROM sistema.proveedor
		WHERE id = $1
	`

	InsertProveedor = `
		INSERT INTO sistema.proveedor (nombre, codigo, descripcion, is_active, tenant_id, razon_i)
		VALUES ($1, $2, $3, $4, current_setting('app.tenant_id', true)::uuid, 'Creacion')
		RETURNING id
	`

	UpdateProveedor = `
		UPDATE sistema.proveedor
		SET nombre = $1, codigo = $2, descripcion = $3, is_active = $4, razon_m = 'Actualizacion'
		WHERE id = $5
	`

	DeleteProveedor = `
		DELETE FROM sistema.proveedor
		WHERE id = $1
	`
)
