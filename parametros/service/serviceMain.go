package service

import (
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/parametros"
	"PlantillaInstitucional/shared/db"
	reporte "PlantillaInstitucional/shared/reportes"
	"context"
	"fmt"
)

func (s *ParametrosService) ListarParametros(ctx context.Context) (pagination.CriteriosPagina[parametros.Parametro], error) {
	criterios := pagination.NewCriteriosPaginaDefault[parametros.Parametro]()
	return s.ListarParametrosFiltrados(ctx, criterios)
}

func (s *ParametrosService) ListarParametrosFiltrados(ctx context.Context, criterios pagination.CriteriosPagina[parametros.Parametro]) (pagination.CriteriosPagina[parametros.Parametro], error) {

	parametross, count, err := s.repo.ConsultarParametros(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "falló al consultar parametross desde el repositorio", err, "operation", "ListarParametrosFiltrados")
		return pagination.CriteriosPagina[parametros.Parametro]{}, fmt.Errorf("error listando parametross: %w", err)
	}

	criterios.Data = parametross
	criterios.Count = count

	s.logger.InfoContext(ctx, "parametross listados correctamente", "total_parametross", count, "operation", "ListarParametrosFiltrados")
	return criterios, nil
}

func (s *ParametrosService) ObtenerParametro(ctx context.Context, idParam string) (parametros.Parametro, error) {
	id, err := db.ParseUUID(idParam)
	if err != nil {
		////s.logger.Error("ID de parametros inválido recibido", "id_param", idParam, "error", err, "operation", "ObtenerParametro")
		return parametros.Parametro{}, fmt.Errorf("ID inválido: %w", err)
	}

	respuestaParametro, err := s.repo.ConsultarParametroPorID(ctx, id)
	if err != nil {
		////s.logger.Error("falló al consultar parametros por ID", "parametros_id", idParam, "error", err, "operation", "ObtenerParametro")
		return parametros.Parametro{}, fmt.Errorf("error obteniendo parametros con ID %s: %w", idParam, err)
	}

	////s.logger.Info("parametros obtenido correctamente", "parametros_id", id, "operation", "ObtenerParametro")
	return respuestaParametro, nil
}

func (s *ParametrosService) CrearParametro(ctx context.Context, parametros parametros.Parametro) error {

	if err := s.validate.Struct(parametros); err != nil {
		s.logger.ErrorContext(ctx, "falló la validación de la estructura Parametro", err, "operation", "CrearParametro")
		return err
	}

	err := s.repo.CrearParametro(ctx, parametros)
	if err != nil {
		s.logger.ErrorContext(ctx, "falló la persistencia del nuevo parametros en el repositorio", err, "parametros_id", parametros.ParametroId, "parametros_key", parametros.Clave, "operation", "CrearParametro")
		return fmt.Errorf("error persistiendo parametros: %w", err)
	}

	s.logger.InfoContext(ctx, "parametros creado correctamente", "parametros_id", parametros.ParametroId, "operation", "CrearParametro")
	return nil
}

func (s *ParametrosService) ActualizarParametro(ctx context.Context, parametros parametros.Parametro) error {

	err := s.validate.Struct(parametros)
	if err != nil {
		////s.logger.Error("falló la validación de la estructura Parametro para actualización", "error", err, "parametros_id", parametros.Id, "operation", "ActualizarParametro")
		return err
	}

	err = s.repo.ActualizarParametro(ctx, parametros)
	if err != nil {
		////s.logger.Error("falló la actualización del parametros en el repositorio", "parametros_id", parametros.Id, "error", err, "operation", "ActualizarParametro")
		return fmt.Errorf("error actualizando Parametro con ID %s: %w", parametros.ParametroId, err)
	}

	////s.logger.Info("parametros actualizado correctamente", "parametros_id", parametros.Id, "operation", "ActualizarParametro")
	return nil
}

func (s *ParametrosService) EliminarParametro(ctx context.Context, idParam string) error {

	id, err := db.ParseUUID(idParam)
	if err != nil {
		////s.logger.Error("ID de Parametro inválido para eliminación", "id_param", idParam, "error", err, "operation", "EliminarParametro")
		return fmt.Errorf("ID inválido: %w", err)
	}

	err = s.repo.EliminarParametro(ctx, id)
	if err != nil {
		////s.logger.Error("falló la eliminación del parametros en el repositorio", "parametros_id", id, "error", err, "operation", "EliminarParametro")
		return fmt.Errorf("error eliminando parametros con ID %s: %w", idParam, err)
	}

	////s.logger.Info("Parametro eliminado correctamente", "parametros_id", id, "operation", "EliminarParametro")
	return nil
}

func (s *ParametrosService) ReportePdf(ctx context.Context) (string, error) {

	criterios := pagination.NewCriteriosPaginaDefault[parametros.Parametro]()
	criterios.Size = reporte.MaxReportRows + 1
	// Ejecutar la consulta
	usuarios, total, err := s.repo.ConsultarParametros(ctx, criterios)
	if err != nil {
		////s.logger.Error("falló al consultar usuarios para reporte PDF", "error", err, "operation", "ReportePdf")
		return "", fmt.Errorf("error al obtener usuarios para reporte PDF: %w", err)
	}
	if total > reporte.MaxReportRows || len(usuarios) > reporte.MaxReportRows {
		return "", reporte.ErrReportRowLimit
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		usuarios, // Datos
		[]string{"Clave", "Valor", "Descripcion", "TipoDato", "IsActive"}, // Encabezados
		[]string{"Clave", "Valor", "Descripcion", "TipoDato", "IsActive"}, // Rutas de los campos
		"reporte_parametross",       // Nombre del archivo
		"Reporte de parametross ",   // Título
		map[string]func(any) string{ // Funciones de renderizado personalizadas
		},
		true,
	)

	if err != nil {
		////s.logger.Error("falló la generación del reporte PDF", "error", err, "operation", "ReportePdf")
		return "", fmt.Errorf("error generando el PDF: %w", err)
	}

	// Retornar la ruta del archivo generado
	////s.logger.Info("reporte PDF generado correctamente", "ruta_archivo", filePath, "operation", "ReportePdf")
	return url, nil
}

func (s *ParametrosService) ReporteExcel(ctx context.Context) (string, error) {
	criterios := pagination.NewCriteriosPaginaDefault[parametros.Parametro]()
	criterios.Size = reporte.MaxReportRows + 1
	// Ejecutar la consulta
	usuarios, total, err := s.repo.ConsultarParametros(ctx, criterios)
	if err != nil {
		////s.logger.Error("falló al consultar Parametros para reporte Excel", "error", err, "operation", "ReporteExcel")
		return "", fmt.Errorf("error al obtener Parametros para reporte Excel: %w", err)
	}
	if total > reporte.MaxReportRows || len(usuarios) > reporte.MaxReportRows {
		return "", reporte.ErrReportRowLimit
	}
	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		usuarios, // Datos
		[]string{"Clave", "Valor", "Descripcion", "TipoDato", "IsActive"}, // Encabezados
		[]string{"Clave", "Valor", "Descripcion", "TipoDato", "IsActive"}, // Rutas de los campos
		"reporte_parametross",       // Nombre del archivo
		"Reporte de Parametros",     // Título
		map[string]func(any) string{ // Funciones de renderizado personalizadas
		},
		true,
	)

	if err != nil {
		////s.logger.Error("falló la generación del reporte Excel", "error", err, "operation", "ReporteExcel")
		return "", fmt.Errorf("error generando el Excel: %w", err)
	}

	// Retornar la ruta del archivo generado
	////s.logger.Info("reporte Excel generado correctamente", "ruta_archivo", filePath, "operation", "ReporteExcel")
	return url, nil
}
