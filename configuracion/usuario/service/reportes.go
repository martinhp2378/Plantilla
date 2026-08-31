package service

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/pagination"
	reporte "PlantillaInstitucional/shared/reportes"
	"context"
	"fmt"
)

// Constantes para encabezados aseguran consistencia entre PDF y Excel
var (
	reportHeaders = []string{"Nombre", "Email", "Perfil"}
	reportFields  = []string{"NombreCompleto", "Email", "Perfil"}
)

func (s *UsuarioService) ReportePdf(ctx context.Context) (string, error) {
	// Advertencia: Obtener todos los registros (99999) consume mucha memoria.
	// Para sistemas grandes, considera "streaming" o paginación por lotes.
	data, err := s.fetchAllUsuarios(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		data,
		reportHeaders,
		reportFields,
		"reporte_usuarios",
		"Reporte de Usuarios",
		nil, // Sin funciones custom
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación PDF", err, "op", "ReportePdf")
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	return url, nil
}

func (s *UsuarioService) ReporteExcel(ctx context.Context) (string, error) {
	data, err := s.fetchAllUsuarios(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		data,
		reportHeaders,
		reportFields,
		"reporte_usuarios",
		"Reporte de Usuarios",
		nil,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación Excel", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error generando Excel: %w", err)
	}

	return url, nil
}

// Helper privado para reutilizar la lógica de extracción masiva
func (s *UsuarioService) fetchAllUsuarios(ctx context.Context) ([]usuario.Usuario, error) {
	criterios := pagination.NewCriteriosPaginaDefault[usuario.Usuario]()
	criterios.Size = 5000
	criterios.First = 0

	var usuariosValores []usuario.Usuario

	for {
		usuariosList, total, err := s.repo.ConsultarUsuarios(ctx, criterios)
		if err != nil {
			s.logger.ErrorContext(ctx, "Error obteniendo datos para reporte", err, "op", "fetchAllUsuarios")
			return nil, fmt.Errorf("error obteniendo datos: %w", err)
		}

		usuariosValores = append(usuariosValores, usuariosList...)

		criterios.First += criterios.Size
		if criterios.First >= total {
			break
		}
	}
	return usuariosValores, nil
}
