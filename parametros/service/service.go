package service

import (
	"context"
)

// LoadParameters loads all parameters from the repository.

// GetParameter retrieves a parameter by its key.
func (s *ParametrosService) GetParameter(ctx context.Context, key string) string {
	val, err := s.repo.GetJSONParamContext(ctx, key)
	if err != nil {
		return ""
	}
	return val
}

// GetParameterContext retrieves a parameter by key using the request context,
// allowing the database search_path to resolve tenant-specific parameters.

// StartParameterMonitoring starts monitoring for parameter changes.
