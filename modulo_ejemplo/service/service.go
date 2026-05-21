package service

import (
	"PlantillaInstitucional/modulo_ejemplo"
	"context"
)

type exampleService struct {
	repo modulo_ejemplo.Repository
}

func NewExampleService(repo modulo_ejemplo.Repository) modulo_ejemplo.Service {
	return &exampleService{repo: repo}
}

func (s *exampleService) ListExamples(ctx context.Context) ([]*modulo_ejemplo.Example, error) {
	return s.repo.GetAll(ctx)
}

func (s *exampleService) GetExample(ctx context.Context, id int64) (*modulo_ejemplo.Example, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *exampleService) CreateExample(ctx context.Context, name string) error {
	e := &modulo_ejemplo.Example{Name: name}
	return s.repo.Create(ctx, e)
}
