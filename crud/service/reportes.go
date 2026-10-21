package service

import (
	reporte "PlantillaInstitucional/shared/reportes"
	"context"
	"fmt"
)

var (
	crudReportHeaders = []string{"Texto Corto", "Entero", "Decimal", "Booleano", "Fecha", "Hora", "Selección", "Correo"}
	crudReportFields  = []string{"TextoCorto", "Entero", "Decimal", "Booleano", "Fecha", "Hora", "Seleccion", "Correo"}
)

func (s *CrudService) ReportePdf(ctx context.Context) (string, error) {
	data, err := s.ObtenerTodos(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		data,
		crudReportHeaders,
		crudReportFields,
		"reporte_crud",
		"Reporte de Registros CRUD",
		nil,
		true,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación PDF crud", err, "op", "ReportePdf")
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	return url, nil
}

func (s *CrudService) ReporteExcel(ctx context.Context) (string, error) {
	data, err := s.ObtenerTodos(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		data,
		crudReportHeaders,
		crudReportFields,
		"reporte_crud",
		"Reporte de Registros CRUD",
		nil,
		true,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación Excel crud", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error generando Excel: %w", err)
	}

	return url, nil
}
