package service

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	reporte "PlantillaInstitucional/shared/reportes"
	"context"
	"fmt"
)

// Constantes para encabezados aseguran consistencia entre PDF y Excel
var (
	reportHeaders = []string{"Nombre", "Apellido Paterno", "Apellido Materno", "RFC", "Curp", "Clave Issemym"}
	reportFields  = []string{"Nombre", "Apellido1", "Apellido2", "Rfc", "Curp", "ClaveIssemym"}
)

func (s *PersonaService) ReportePdf(ctx context.Context) (string, error) {
	// Advertencia: Obtener todos los registros (99999) consume mucha memoria.
	// Para sistemas grandes, considera "streaming" o paginación por lotes.
	data, err := s.fetchAllPersonas(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		data,
		reportHeaders,
		reportFields,
		"reporte_identidades",
		"Reporte de Personas",
		nil,
		true,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación PDF", err, "op", "ReportePdf")
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	return url, nil
}

func (s *PersonaService) ReporteExcel(ctx context.Context) (string, error) {
	data, err := s.fetchAllPersonas(ctx)
	if err != nil {
		return "", err
	}

	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		data,
		reportHeaders,
		reportFields,
		"reporte_identidades",
		"Reporte de Personas",
		nil,
		true,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación Excel", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error generando Excel: %w", err)
	}

	return url, nil
}

// Helper privado para reutilizar la lógica de extracción masiva
func (s *PersonaService) fetchAllPersonas(ctx context.Context) ([]institucion.Persona, error) {
	criterios := pagination.NewCriteriosPaginaDefault[institucion.Persona]()
	criterios.Size = 5000
	criterios.First = 0

	var personasValores []institucion.Persona

	for {
		personasList, total, err := s.repo.ConsultarPersonas(ctx, criterios)
		if err != nil {
			s.logger.ErrorContext(ctx, "Error obteniendo datos para reporte", err, "op", "fetchAllPersonas")
			return nil, fmt.Errorf("error obteniendo datos: %w", err)
		}

		for _, p := range personasList {
			if p != nil {
				personasValores = append(personasValores, *p)
			}
		}

		criterios.First += criterios.Size
		if criterios.First >= total {
			break
		}
	}

	return personasValores, nil
}
