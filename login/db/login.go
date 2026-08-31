package db

import (
	"context"
	"errors"
	"fmt"

	"PlantillaInstitucional/login"
	errores "PlantillaInstitucional/shared/errores"
	pass "PlantillaInstitucional/shared/password"

	"github.com/issemym-repos/sesiones"

	"github.com/jackc/pgx/v5"
)

var ErrProfilePolicyInvalid = errors.New("profile policy is invalid")

func validateProfilePolicyForLogin(ctx context.Context, resolver PolicyResolver, profile string) error {
	if resolver == nil {
		return ErrProfilePolicyInvalid
	}

	active, err := resolver.IsProfileActive(ctx, profile)
	if err != nil {
		return errors.Join(ErrProfilePolicyInvalid, err)
	}
	if !active {
		return ErrProfilePolicyInvalid
	}

	return nil
}

func (r RepositorioPsql) ValidarEmailPassword(ctx context.Context,
	email string, password string) (*login.DatosSesion, error) {

	var row RowUsuario
	
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
		select uuid_persona,
			password,
			email,
			perfil
		from sesion.usuario
		where email = $1`, email).Scan(&row.UuidPersona, &row.Password, &row.Email, &row.Perfil)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Join(err, errores.ErrRegistroNoEncontrado)
		}
		return nil, fmt.Errorf("error al obtener el usuario: %w", err)
	}

	// Usar CheckPasswordHash para comparar de forma segura
	if !pass.CheckPasswordHash(password, row.Password) {
		return nil, errores.ErrCredencialesInvalidas
	}

	if err := validateProfilePolicyForLogin(ctx, r.PolicyResolver, row.Perfil); err != nil {
		return nil, errors.Join(errores.ErrCredencialesInvalidas, err)
	}

	return &login.DatosSesion{
		UuidPersona: row.UuidPersona,
		Email:       row.Email,
		Perfil:      sesiones.Perfil(row.Perfil),
	}, nil
}

