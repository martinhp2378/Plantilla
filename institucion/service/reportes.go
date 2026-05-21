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
	criterios.Size = 5000 // Limite duro de seguridad, ajusta según memoria disponible
	criterios.First = 0

	// Nota: Si hay más de 5000, aquí solo traemos los primeros.
	// Lo ideal sería iterar páginas hasta traer todo si el reporte debe ser exhaustivo.
	personasList, _, err := s.repo.ConsultarPersonas(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error obteniendo datos para reporte", err, "op", "fetchAllPersonas")
		return nil, fmt.Errorf("error obteniendo datos: %w", err)
	}

	var personasValores []institucion.Persona
	for _, p := range personasList {
		if p != nil {
			personasValores = append(personasValores, *p)
		}
	}

	return personasValores, nil
}
