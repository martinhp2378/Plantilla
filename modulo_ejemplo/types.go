package modulo_ejemplo

import (
	"context"
	"net/http"
)

type Example struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Repository interface {
	GetAll(ctx context.Context) ([]*Example, error)
	GetByID(ctx context.Context, id int64) (*Example, error)
	Create(ctx context.Context, e *Example) error
}

type Service interface {
	ListExamples(ctx context.Context) ([]*Example, error)
	GetExample(ctx context.Context, id int64) (*Example, error)
	CreateExample(ctx context.Context, name string) error
}

type Handler interface {
	RegisterRoutes(mux http.Handler)
	HandleList(w http.ResponseWriter, r *http.Request)
	HandleCreate(w http.ResponseWriter, r *http.Request)
}
