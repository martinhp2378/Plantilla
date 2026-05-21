package sesiones

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// ConstructorSesionJWT es un constructor especializado para manejar sesiones JWT.
// Utiliza el ConstructorSesion base y agrega funcionalidad específica de JWT.
type ConstructorSesionJWT struct {
	*ConstructorSesion
	configJWT *ConfigJWT
}

// NuevoConstructorJWT crea un nuevo constructor de sesiones JWT con la
// configuración proporcionada, si se habilita Redis se utiliza una
// configuración default; para especificarla, utilizar
// [NuevoConstructorJWTConRedis].
func NuevoConstructorJWT(conexion PgxQueryable,
	config *ConfigSesion,
	configJWT *ConfigJWT,
	registroTipos *RegistroTipos,
	fnObtenerDatos func(uuid.UUID, Perfil, ...map[string]any) []any) (*ConstructorSesionJWT, error) {

	if err := ValidarConfigJWT(configJWT); err != nil {
		return nil, fmt.Errorf("configuración JWT inválida: %w", err)
	}

	cs := NuevoConstructor(conexion, config, registroTipos, fnObtenerDatos)

	return &ConstructorSesionJWT{
		ConstructorSesion: cs,
		configJWT:         configJWT,
	}, nil
}

// NuevoConstructorJWTConRedis crea un nuevo constructor de sesiones JWT con el
// cliente de Redis proveído.
func NuevoConstructorJWTConRedis(conexion PgxQueryable,
	config *ConfigSesion,
	configJWT *ConfigJWT,
	registroTipos *RegistroTipos,
	fnObtenerDatos func(uuid.UUID, Perfil, ...map[string]any) []any,
	redisClient *redis.Client) (*ConstructorSesionJWT, error) {

	if err := ValidarConfigJWT(configJWT); err != nil {
		return nil, fmt.Errorf("configuración JWT inválida: %w", err)
	}

	cs := NuevoConstructorConRedis(conexion, config, registroTipos, fnObtenerDatos, redisClient)

	return &ConstructorSesionJWT{
		ConstructorSesion: cs,
		configJWT:         configJWT,
	}, nil
}

// GenerarAccessToken genera un token de acceso JWT para una sesión.
func (cj *ConstructorSesionJWT) GenerarAccessToken(sesion *Sesion) (string, error) {
	ahora := time.Now()
	expiracion := ahora.Add(cj.configJWT.AccessTokenExpiry)

	uuidPersonaStr := ""
	if sesion.UuidPersona != nil {
		uuidPersonaStr = sesion.UuidPersona.String()
	}

	claims := &SesionClaims{
		UuidSesion:  sesion.UuidSesion.String(),
		UuidPersona: uuidPersonaStr,
		Perfil:      string(sesion.Perfil),
		TipoToken:   TipoTokenAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracion),
			IssuedAt:  jwt.NewNumericDate(ahora),
			NotBefore: jwt.NewNumericDate(ahora),
			Issuer:    cj.configJWT.Issuer,
			Audience:  cj.configJWT.Audience,
		},
	}

	var token *jwt.Token
	switch cj.configJWT.Algorithm {
	case "HS256":
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	case "HS384":
		token = jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	case "HS512":
		token = jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	default:
		return "", fmt.Errorf("algoritmo JWT no soportado: %s", cj.configJWT.Algorithm)
	}

	tokenString, err := token.SignedString(cj.configJWT.SecretKey)
	if err != nil {
		return "", ErrorTokenInvalido{err: fmt.Errorf("error al firmar token: %w", err)}
	}

	return tokenString, nil
}

// GenerarRefreshToken genera un refresh token JWT para una sesión.
// Este token tiene una expiración más larga y se usa para obtener nuevos access tokens.
func (cj *ConstructorSesionJWT) GenerarRefreshToken(sesion *Sesion) (string, error) {
	if !cj.configJWT.EnableRefreshTokens {
		return "", fmt.Errorf("refresh tokens no están habilitados en la configuración")
	}

	ahora := time.Now()
	expiracion := ahora.Add(cj.configJWT.RefreshTokenExpiry)

	uuidPersonaStr := ""
	if sesion.UuidPersona != nil {
		uuidPersonaStr = sesion.UuidPersona.String()
	}

	claims := &SesionClaims{
		UuidSesion:  sesion.UuidSesion.String(),
		UuidPersona: uuidPersonaStr,
		Perfil:      string(sesion.Perfil),
		TipoToken:   TipoTokenRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracion),
			IssuedAt:  jwt.NewNumericDate(ahora),
			NotBefore: jwt.NewNumericDate(ahora),
			Issuer:    cj.configJWT.Issuer,
			Audience:  cj.configJWT.Audience,
		},
	}

	// Crear token con el algoritmo especificado
	var token *jwt.Token
	switch cj.configJWT.Algorithm {
	case "HS256":
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	case "HS384":
		token = jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	case "HS512":
		token = jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	default:
		return "", fmt.Errorf("algoritmo JWT no soportado: %s", cj.configJWT.Algorithm)
	}

	// Firmar token
	tokenString, err := token.SignedString(cj.configJWT.SecretKey)
	if err != nil {
		return "", ErrorRefreshTokenInvalido{err: fmt.Errorf("error al firmar refresh token: %w", err)}
	}

	// Guardar hash del refresh token en la sesión para validación posterior
	hash := sha256.Sum256([]byte(tokenString))
	sesion.RefreshTokenHash = hash[:]

	return tokenString, nil
}

// GenerarParTokens genera un par de access token y refresh token.
func (cj *ConstructorSesionJWT) GenerarParTokens(sesion *Sesion) (accessToken, refreshToken string, err error) {
	accessToken, err = cj.GenerarAccessToken(sesion)
	if err != nil {
		return "", "", err
	}

	if cj.configJWT.EnableRefreshTokens {
		refreshToken, err = cj.GenerarRefreshToken(sesion)
		if err != nil {
			return "", "", err
		}
	}

	return accessToken, refreshToken, nil
}

// ValidarToken valida un token JWT y regresa los claims si es válido.
func (cj *ConstructorSesionJWT) ValidarToken(tokenString string) (*SesionClaims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{cj.configJWT.Algorithm}),
		jwt.WithIssuer(cj.configJWT.Issuer),
		jwt.WithAudience(cj.configJWT.Audience...),
		jwt.WithLeeway(cj.configJWT.AllowedClockSkew),
	)

	token, err := parser.ParseWithClaims(tokenString, &SesionClaims{}, func(token *jwt.Token) (interface{}, error) {
		return cj.configJWT.SecretKey, nil
	})

	if err != nil {
		// Distinguir entre token expirado y otros errores
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrorTokenExpirado{err: err}
		}
		return nil, ErrorTokenInvalido{err: err}
	}

	claims, ok := token.Claims.(*SesionClaims)
	if !ok {
		return nil, ErrorTokenInvalido{err: fmt.Errorf("claims inválidos")}
	}

	return claims, nil
}

// ValidarRefreshToken valida un refresh token y verifica que coincida con el hash almacenado.
func (cj *ConstructorSesionJWT) ValidarRefreshToken(tokenString string, sesion *Sesion) error {
	claims, err := cj.ValidarToken(tokenString)
	if err != nil {
		return ErrorRefreshTokenInvalido{err: err}
	}

	if claims.TipoToken != TipoTokenRefresh {
		return ErrorRefreshTokenInvalido{err: fmt.Errorf("token no es un refresh token")}
	}

	if claims.UuidSesion != sesion.UuidSesion.String() {
		return ErrorRefreshTokenInvalido{err: fmt.Errorf("UUID de sesión no coincide")}
	}

	hash := sha256.Sum256([]byte(tokenString))
	hashCalculado := hash[:]

	if len(sesion.RefreshTokenHash) == 0 {
		return ErrorRefreshTokenInvalido{err: fmt.Errorf("no hay refresh token hash almacenado")}
	}

	if len(hashCalculado) != len(sesion.RefreshTokenHash) {
		return ErrorRefreshTokenInvalido{err: fmt.Errorf("hash no coincide")}
	}
	for i := range hashCalculado {
		if hashCalculado[i] != sesion.RefreshTokenHash[i] {
			return ErrorRefreshTokenInvalido{err: fmt.Errorf("hash no coincide")}
		}
	}

	return nil
}

// extraerTokenDeHeader extrae el token JWT del header Authorization.
// Espera el formato: "Authorization: Bearer <token>"
func extraerTokenDeHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrorTokenNoEncontrado{detalle: "header Authorization vacío"}
	}

	// Formato esperado: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrorTokenNoEncontrado{detalle: "formato inválido, se esperaba 'Bearer <token>'"}
	}

	return parts[1], nil
}

// FirmarPersonaJWT firma una sesión con JWT, asignando el usuario y generando tokens.
func (cj *ConstructorSesionJWT) FirmarPersonaJWT(ctx context.Context, sesion *Sesion,
	uuidPersona uuid.UUID, perfil Perfil, datosExtra ...map[string]any) (accessToken, refreshToken string, err error) {

	err = cj.FirmarPersona(sesion, uuidPersona, perfil, datosExtra...)
	if err != nil {
		return "", "", err
	}

	sesion.TipoSesion = TipoSesionJWT

	accessToken, refreshToken, err = cj.GenerarParTokens(sesion)
	if err != nil {
		sesion.CerrarSesion()
		return "", "", err
	}

	err = cj.AlmacenarSesion(ctx, sesion)
	if err != nil {
		sesion.CerrarSesion()
		return "", "", fmt.Errorf("error al almacenar sesión: %w", err)
	}

	return accessToken, refreshToken, nil
}

// RefrescarTokenJWT genera un nuevo access token a partir de un refresh token válido.
func (cj *ConstructorSesionJWT) RefrescarTokenJWT(ctx context.Context, refreshToken string) (nuevoAccessToken string, err error) {
	if !cj.configJWT.EnableRefreshTokens {
		return "", fmt.Errorf("refresh tokens no están habilitados")
	}

	// Recuperar la sesión a partir del refresh token

	claims, err := cj.ValidarToken(refreshToken)
	if err != nil {
		return "", ErrorRefreshTokenInvalido{err: err}
	}

	if claims.TipoToken != TipoTokenRefresh {
		return "", ErrorRefreshTokenInvalido{err: fmt.Errorf("token no es un refresh token")}
	}

	uuidSesion, err := uuid.FromString(claims.UuidSesion)
	if err != nil {
		return "", ErrorRefreshTokenInvalido{err: fmt.Errorf("UUID de sesión inválido: %w", err)}
	}

	sesion, err := cj.SesionDeUuid(ctx, uuidSesion)
	if err != nil {
		return "", ErrorRefreshTokenInvalido{err: fmt.Errorf("sesión no encontrada: %w", err)}
	}

	if sesion.TipoSesion != TipoSesionJWT {
		return "", ErrorRefreshTokenInvalido{err: fmt.Errorf("sesión no es de tipo JWT")}
	}

	// Validar el refresh token contra el hash almacenado

	err = cj.ValidarRefreshToken(refreshToken, sesion)
	if err != nil {
		return "", err
	}

	// Verificar que la sesión no haya expirado, si ya expiró se elimina para
	// evitar fugas de memoria
	if cj.YaExpiro(sesion) {
		_ = cj.EliminarSesion(ctx, sesion.UuidSesion)
		return "", ErrorTokenExpirado{err: fmt.Errorf("sesión expirada")}
	}

	// Actualizar timestamp de último acceso
	cj.Tocar(sesion)

	// Generar nuevo access token
	nuevoAccessToken, err = cj.GenerarAccessToken(sesion)
	if err != nil {
		return "", err
	}

	// Actualizar sesión en el store
	err = cj.ActualizarSesion(ctx, sesion)
	if err != nil {
		return "", fmt.Errorf("error al actualizar sesión: %w", err)
	}

	return nuevoAccessToken, nil
}

// RevocarSesionJWT revoca una sesión JWT eliminándola del store.
// Después de esto, tanto el access token como el refresh token dejarán de funcionar.
func (cj *ConstructorSesionJWT) RevocarSesionJWT(ctx context.Context, sesion *Sesion) error {
	if sesion.TipoSesion != TipoSesionJWT {
		return fmt.Errorf("la sesión no es de tipo JWT")
	}

	if sesion.UuidPersona == nil {
		return fmt.Errorf("la sesión no tiene UUID de persona")
	}

	// Guardar UUID de persona antes de cerrar la sesión
	uuidPersona := *sesion.UuidPersona

	// Limpiar refresh token hash
	sesion.RefreshTokenHash = nil

	// Cerrar la sesión (limpia datos)
	sesion.CerrarSesion()

	// Eliminar sesiones de la persona del store
	return cj.EliminarSesionesPersona(ctx, uuidPersona)
}
