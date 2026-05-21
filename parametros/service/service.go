package service

import (
	"context"
	"fmt"
)

// LoadParameters loads all parameters from the repository.
func (s *ParametrosService) LoadParameters() error {
	err := s.repo.CargarParametros()
	if err != nil {
		return fmt.Errorf("error loading parameters: %w", err)
	}
	return nil
}

// GetParameter retrieves a parameter by its key.
func (s *ParametrosService) GetParameter(key string) string {
	val, err := s.repo.GetJSONParam(key)
	if err != nil {
		return ""
	}
	return val
}

// GetParameterContext retrieves a parameter by key using the request context,
// allowing the database search_path to resolve tenant-specific parameters.
func (s *ParametrosService) GetParameterContext(ctx context.Context, key string) string {
	val, err := s.repo.GetJSONParamContext(ctx, key)
	if err != nil {
		return ""
	}
	return val
}

// StartParameterMonitoring starts monitoring for parameter changes.
func (s *ParametrosService) StartParameterMonitoring(updateChan <-chan string) {
	s.repo.MonitorParametros(updateChan)
}
