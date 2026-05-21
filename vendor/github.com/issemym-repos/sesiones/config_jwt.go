package sesiones

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ConfigJWT struct {
	SecretKey           []byte
	Algorithm           string        // "HS256", "RS256", etc.
	AccessTokenExpiry   time.Duration // e.g., 15 minutes
	RefreshTokenExpiry  time.Duration // e.g., 7 days
	EnableRefreshTokens bool
	Issuer              string        // JWT "iss" claim
	Audience            []string      // JWT "aud" claim
	AllowedClockSkew    time.Duration // Tolerancia para validar el access token
}

func ConfigJWTDefault() *ConfigJWT {
	return &ConfigJWT{
		SecretKey:           []byte("CAMBIAR_PARA_PRODUCCION"), // Generar una llave en producción
		Algorithm:           "HS256",
		AccessTokenExpiry:   15 * time.Minute,
		RefreshTokenExpiry:  7 * 24 * time.Hour, // 7 días
		EnableRefreshTokens: true,
		Issuer:              "libreria-sesiones",
		Audience:            []string{"mi-api-backend"},
		AllowedClockSkew:    30 * time.Second,
	}
}

// ValidarConfigJWT valida que la configuración de JWT sea válida y segura.
func ValidarConfigJWT(cfg *ConfigJWT) error {
	if cfg == nil {
		return ErrorPunteroNulo
	}

	if len(cfg.SecretKey) == 0 {
		return fmt.Errorf("ConfigJWT.SecretKey no puede estar vacía")
	}

	// Obligar mínimo 32 bytes por seguridad
	if len(cfg.SecretKey) < 32 {
		return fmt.Errorf("ConfigJWT.SecretKey debe tener al menos 32 bytes, tiene %d", len(cfg.SecretKey))
	}

	// Validar algoritmo
	algoritmosPermitidos := []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512"}
	algoritmoValido := false
	for _, alg := range algoritmosPermitidos {
		if cfg.Algorithm == alg {
			algoritmoValido = true
			break
		}
	}
	if !algoritmoValido {
		return fmt.Errorf("ConfigJWT.Algorithm '%s' no es válido. Debe ser uno de: %v", cfg.Algorithm, algoritmosPermitidos)
	}

	// Validar que access token expiry sea razonable (entre 1 minuto y 24 horas)
	if cfg.AccessTokenExpiry < 1*time.Minute {
		return fmt.Errorf("ConfigJWT.AccessTokenExpiry debe ser al menos 1 minuto")
	}
	if cfg.AccessTokenExpiry > 24*time.Hour {
		return fmt.Errorf("ConfigJWT.AccessTokenExpiry no debe ser mayor a 24 horas por seguridad")
	}

	// Validar refresh token expiry
	if cfg.EnableRefreshTokens {
		if cfg.RefreshTokenExpiry < 1*time.Hour {
			return fmt.Errorf("ConfigJWT.RefreshTokenExpiry debe ser al menos 1 hora")
		}
		if cfg.RefreshTokenExpiry > 30*24*time.Hour {
			return fmt.Errorf("ConfigJWT.RefreshTokenExpiry no debe ser mayor a 30 días")
		}
	}

	if cfg.Issuer == "" {
		return fmt.Errorf("ConfigJWT.Issuer no puede estar vacío")
	}

	if len(cfg.Audience) == 0 {
		return fmt.Errorf("ConfigJWT.Audience debe tener al menos un valor")
	}

	return nil
}

type TipoToken string

const (
	TipoTokenAccess  TipoToken = "access"
	TipoTokenRefresh TipoToken = "refresh"
)

type SesionClaims struct {
	UuidSesion  string    `json:"uuid_sesion"`
	UuidPersona string    `json:"uuid_persona,omitempty"`
	Perfil      string    `json:"perfil"`
	TipoToken   TipoToken `json:"tipo_token"`
	jwt.RegisteredClaims
}
