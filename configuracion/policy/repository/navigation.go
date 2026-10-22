package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// NavigationPaths batches the global effective consultar grants for one profile.
func (r *PolicyRepository) NavigationPaths(ctx context.Context, profile string) ([]string, error) {
	var paths []string
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, navigationPathsQuery, profile)
		if err != nil {
			return err
		}
		var collectErr error
		paths, collectErr = pgx.CollectRows(rows, pgx.RowTo[string])
		return collectErr
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}
