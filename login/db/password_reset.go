package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidOrExpiredResetToken = errors.New("invalid or expired reset token")

// IssuePasswordReset replaces any earlier unused token for the account. The
// caller must always return the same response whether this returns false or true.
func (r RepositorioPsql) IssuePasswordReset(ctx context.Context, email string, tokenHash []byte, expiresAt time.Time) (bool, error) {
	found := false
	err := r.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var personID string
		if err := tx.QueryRow(ctx, `SELECT uuid_persona FROM sesion.usuario WHERE email = $1`, email).Scan(&personID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("buscar cuenta para recuperación: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM sesion.password_reset_token WHERE uuid_persona = $1 AND used_at IS NULL`, personID); err != nil {
			return fmt.Errorf("invalidar token anterior: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sesion.password_reset_token (uuid_persona, token_hash, expires_at, tenant_id) VALUES ($1, $2, $3, current_setting('app.tenant_id', true)::uuid)`, personID, tokenHash, expiresAt); err != nil {
			return fmt.Errorf("guardar token de recuperación: %w", err)
		}
		found = true
		return nil
	})
	return found, err
}

// ConsumePasswordReset atomically consumes a valid token and enables the account
// with the caller-supplied password hash. A consumed token cannot be replayed.
func (r RepositorioPsql) ConsumePasswordReset(ctx context.Context, tokenHash []byte, passwordHash string) error {
	return r.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var personID string
		if err := tx.QueryRow(ctx, `
			DELETE FROM sesion.password_reset_token
			WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
			RETURNING uuid_persona`, tokenHash).Scan(&personID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidOrExpiredResetToken
			}
			return fmt.Errorf("consumir token: %w", err)
		}
		ct, err := tx.Exec(ctx, `UPDATE sesion.usuario SET password = $1, is_active = true WHERE uuid_persona = $2`, passwordHash, personID)
		if err != nil {
			return fmt.Errorf("actualizar contraseña: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return ErrInvalidOrExpiredResetToken
		}
		return nil
	})
}
