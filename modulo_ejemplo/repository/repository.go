package repository

import (
	"PlantillaInstitucional/modulo_ejemplo"
	"PlantillaInstitucional/shared/contextkey"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) modulo_ejemplo.Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) getSchema(ctx context.Context) string {
	schema, ok := ctx.Value(contextkey.TenantSchemaKey).(string)
	if !ok || schema == "" {
		return "public"
	}
	return schema
}

func (r *pgRepository) GetAll(ctx context.Context) ([]*modulo_ejemplo.Example, error) {
	schema := r.getSchema(ctx)
	query := fmt.Sprintf("SELECT id, name FROM %s.examples", schema)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var examples []*modulo_ejemplo.Example
	for rows.Next() {
		e := &modulo_ejemplo.Example{}
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, err
		}
		examples = append(examples, e)
	}

	return examples, nil
}

func (r *pgRepository) GetByID(ctx context.Context, id int64) (*modulo_ejemplo.Example, error) {
	schema := r.getSchema(ctx)
	query := fmt.Sprintf("SELECT id, name FROM %s.examples WHERE id = $1", schema)

	e := &modulo_ejemplo.Example{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&e.ID, &e.Name)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *pgRepository) Create(ctx context.Context, e *modulo_ejemplo.Example) error {
	schema := r.getSchema(ctx)
	query := fmt.Sprintf("INSERT INTO %s.examples (name) VALUES ($1) RETURNING id", schema)

	return r.pool.QueryRow(ctx, query, e.Name).Scan(&e.ID)
}
