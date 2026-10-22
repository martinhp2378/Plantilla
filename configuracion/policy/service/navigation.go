package service

import (
	"PlantillaInstitucional/shared/navigation"
	"context"
)

// NavigationPaths is additive: legacy repositories remain valid and fail closed.
func (s *Service) NavigationPaths(ctx context.Context, profile string) ([]string, error) {
	if s == nil {
		return nil, nil
	}
	reader, ok := s.repo.(navigation.Reader)
	if !ok {
		return nil, nil
	}
	return reader.NavigationPaths(ctx, profile)
}
