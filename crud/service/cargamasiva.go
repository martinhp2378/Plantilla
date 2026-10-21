package service

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/shared/cargamasiva"
	"context"
	"errors"
	"fmt"
	"io"
)

// ObtenerConfigCargaCrud define las columnas y reglas para la carga masiva del módulo crud.
func ObtenerConfigCargaCrud() cargamasiva.ConfigCarga[crud.Crud] {
	return cargamasiva.ConfigCarga[crud.Crud]{
		Separador: ',',
		MaxFilas:  5000,
		ColumnasEsperadas: []cargamasiva.DefinicionColumna[crud.Crud]{
			{
				Nombre:      "texto_corto",
				Requerido:   true,
				Descripcion: "Texto descriptivo breve (máximo 100 caracteres)",
				Ejemplo:     "Producto Ejemplo Masivo",
				Parser: func(valor string, c *crud.Crud) error {
					if len(valor) > 100 {
						return errors.New("máximo 100 caracteres permitidos")
					}
					c.TextoCorto = valor
					return nil
				},
			},
			{
				Nombre:      "texto_largo",
				Requerido:   false,
				Descripcion: "Descripción detallada",
				Ejemplo:     "Detalle cargado desde CSV masivo",
				Parser: func(valor string, c *crud.Crud) error {
					c.TextoLargo = valor
					return nil
				},
			},
			{
				Nombre:      "entero",
				Requerido:   true,
				Descripcion: "Número entero",
				Ejemplo:     "50",
				Parser: func(valor string, c *crud.Crud) error {
					n, err := cargamasiva.ParseEntero(valor, 0, 1000000)
					if err != nil {
						return err
					}
					c.Entero = n
					return nil
				},
			},
			{
				Nombre:      "decimal",
				Requerido:   true,
				Descripcion: "Cantidad o precio decimal",
				Ejemplo:     "123.45",
				Parser: func(valor string, c *crud.Crud) error {
					f, err := cargamasiva.ParseDecimal(valor, 0, 10000000)
					if err != nil {
						return err
					}
					c.Decimal = f
					return nil
				},
			},
			{
				Nombre:      "booleano",
				Requerido:   true,
				Descripcion: "Valor booleano ('Sí', 'No', 'true', 'false')",
				Ejemplo:     "Sí",
				Parser: func(valor string, c *crud.Crud) error {
					b, err := cargamasiva.ParseBooleano(valor)
					if err != nil {
						return err
					}
					c.Booleano = b
					return nil
				},
			},
			{
				Nombre:      "fecha",
				Requerido:   true,
				Descripcion: "Fecha en formato AAAA-MM-DD",
				Ejemplo:     "2026-10-21",
				Parser: func(valor string, c *crud.Crud) error {
					f, err := cargamasiva.ParseFecha(valor)
					if err != nil {
						return err
					}
					c.Fecha = f
					return nil
				},
			},
			{
				Nombre:      "hora",
				Requerido:   true,
				Descripcion: "Hora en formato HH:MM (24h)",
				Ejemplo:     "14:30",
				Parser: func(valor string, c *crud.Crud) error {
					h, err := cargamasiva.ParseHora(valor)
					if err != nil {
						return err
					}
					c.Hora = h
					return nil
				},
			},
			{
				Nombre:      "seleccion",
				Requerido:   true,
				Descripcion: "Opción (opcion_a, opcion_b, opcion_c)",
				Ejemplo:     "opcion_a",
				Parser: func(valor string, c *crud.Crud) error {
					s, err := cargamasiva.ParseEnum(valor, []string{"opcion_a", "opcion_b", "opcion_c"})
					if err != nil {
						return err
					}
					c.Seleccion = s
					return nil
				},
			},
			{
				Nombre:      "correo",
				Requerido:   true,
				Descripcion: "Correo electrónico institucional o de contacto",
				Ejemplo:     "usuario.masivo@issemym.gob.mx",
				Parser: func(valor string, c *crud.Crud) error {
					e, err := cargamasiva.ParseEmail(valor)
					if err != nil {
						return err
					}
					c.Correo = e
					return nil
				},
			},
			{
				Nombre:      "telefono",
				Requerido:   false,
				Descripcion: "Teléfono de contacto",
				Ejemplo:     "7221234567",
				Parser: func(valor string, c *crud.Crud) error {
					c.Telefono = valor
					return nil
				},
			},
			{
				Nombre:      "url",
				Requerido:   false,
				Descripcion: "URL válida",
				Ejemplo:     "https://issemym.gob.mx",
				Parser: func(valor string, c *crud.Crud) error {
					u, err := cargamasiva.ParseURL(valor)
					if err != nil {
						return err
					}
					c.URL = u
					return nil
				},
			},
			{
				Nombre:      "color",
				Requerido:   false,
				Descripcion: "Color hexadecimal (ej: #691C32)",
				Ejemplo:     "#691C32",
				Parser: func(valor string, c *crud.Crud) error {
					col, err := cargamasiva.ParseColor(valor)
					if err != nil {
						return err
					}
					c.Color = col
					return nil
				},
			},
		},
	}
}

// GenerarPlantillaCSV exporta el CSV vacío con encabezados y fila ejemplo.
func (s *CrudService) GenerarPlantillaCSV() ([]byte, error) {
	return cargamasiva.GenerarPlantillaCSV(ObtenerConfigCargaCrud())
}

// CargarMasivo procesa el archivo CSV, valida y persiste en lote.
func (s *CrudService) CargarMasivo(ctx context.Context, r io.Reader) (*cargamasiva.ResultadoCarga[crud.Crud], error) {
	cfg := ObtenerConfigCargaCrud()
	resultado, err := cargamasiva.ProcesarCSV(r, cfg)
	if err != nil {
		return nil, err
	}

	// Si hay errores de validación, retornamos el reporte para mostrárselo al usuario
	if !resultado.EsExitoso() {
		return resultado, nil
	}

	// Persistir en base de datos en una sola transacción
	if err := s.repo.CrearCrudMasivo(ctx, resultado.Registros); err != nil {
		return nil, fmt.Errorf("error guardando registros en base de datos: %w", err)
	}

	s.logger.Info("Carga masiva completada exitosamente",
		"total", resultado.TotalValidos,
		"op", "CargarMasivo",
	)

	return resultado, nil
}
