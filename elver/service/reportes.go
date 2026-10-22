package service

import (
	"PlantillaInstitucional/elver"
	reporte "PlantillaInstitucional/shared/reportes"
	"context"
	"fmt"
)

var (
	elverReportHeaders = []string{"Nombre", "Código", "Descripción", "Estatus"}
	elverReportFields  = []string{"Nombre", "Codigo", "Descripcion", "IsActive"}
)

func (s *ElverService) ReportePdf(ctx context.Context) (string, error) {
	data, err := s.repo.ObtenerTodos(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo obteniendo registros para PDF elver", err, "op", "ReportePdf")
		return "", fmt.Errorf("error obteniendo registros: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("no hay registros disponibles para exportar")
	}

	renderFuncs := map[string]func(any) string{
		"IsActive": func(v any) string {
			if item, ok := v.(elver.Elver); ok && item.IsActive {
				return "Activo"
			}
			if item, ok := v.(*elver.Elver); ok && item != nil && item.IsActive {
				return "Activo"
			}
			return "Inactivo"
		},
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		data,
		elverReportHeaders,
		elverReportFields,
		"reporte_elver",
		"Reporte de Elveres",
		renderFuncs,
		true,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación PDF elver", err, "op", "ReportePdf")
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	return url, nil
}

func (s *ElverService) ReporteExcel(ctx context.Context) (string, error) {
	data, err := s.repo.ObtenerTodos(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo obteniendo registros para Excel elver", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error obteniendo registros: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("no hay registros disponibles para exportar")
	}

	renderFuncs := map[string]func(any) string{
		"IsActive": func(v any) string {
			if item, ok := v.(elver.Elver); ok && item.IsActive {
				return "Activo"
			}
			if item, ok := v.(*elver.Elver); ok && item != nil && item.IsActive {
				return "Activo"
			}
			return "Inactivo"
		},
	}

	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		data,
		elverReportHeaders,
		elverReportFields,
		"reporte_elver",
		"Reporte de Elveres",
		renderFuncs,
		true,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación Excel elver", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error generando Excel: %w", err)
	}

	return url, nil
}
