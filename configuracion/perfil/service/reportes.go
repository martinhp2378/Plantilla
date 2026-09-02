package service

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/pagination"
	reporte "PlantillaInstitucional/shared/reportes"
	"context"
	"fmt"
)

// Constantes para encabezados aseguran consistencia entre PDF y Excel
var (
	reportHeaders = []string{"Nombre"}
	reportFields  = []string{"Nombre"}
)

func (s *PerfilService) ReportePdf(ctx context.Context) (string, error) {
	// Fetches in pages and rejects totals above the shared report limit.
	data, err := s.fetchAllPerfiles(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		data,
		reportHeaders,
		reportFields,
		"reporte_perfiles",
		"Reporte de Perfiles",
		nil, // Sin funciones custom
		false,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación PDF", err, "op", "ReportePdf")
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	return url, nil
}

func (s *PerfilService) ReporteExcel(ctx context.Context) (string, error) {
	data, err := s.fetchAllPerfiles(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		data,
		reportHeaders,
		reportFields,
		"reporte_perfiles",
		"Reporte de Perfiles",
		nil,
		false,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación Excel", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error generando Excel: %w", err)
	}

	return url, nil
}

// Helper privado para reutilizar la lógica de extracción masiva
func (s *PerfilService) fetchAllPerfiles(ctx context.Context) ([]perfil.Perfil, error) {
	criterios := pagination.NewCriteriosPaginaDefault[perfil.Perfil]()
	criterios.Size = 5000
	criterios.First = 0

	var perfilesValores []perfil.Perfil

	for {
		perfilesList, total, err := s.repo.ConsultarPerfiles(ctx, criterios)
		if err != nil {
			s.logger.ErrorContext(ctx, "Error obteniendo datos para reporte", err, "op", "fetchAllPerfiles")
			return nil, fmt.Errorf("error obteniendo datos: %w", err)
		}
		if total > reporte.MaxReportRows || len(perfilesValores)+len(perfilesList) > reporte.MaxReportRows {
			return nil, reporte.ErrReportRowLimit
		}

		perfilesValores = append(perfilesValores, perfilesList...)

		criterios.First += criterios.Size
		if criterios.First >= total {
			break
		}
	}
	return perfilesValores, nil
}
