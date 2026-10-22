package service

import (
	"PlantillaInstitucional/miperfil"
	"PlantillaInstitucional/shared/password"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrPasswordActualIncorrecta = errors.New("la contraseña actual no es correcta")
	ErrPasswordNoCoincide       = errors.New("la confirmación de la contraseña no coincide")
	ErrPasswordCorta            = errors.New("la nueva contraseña debe tener al menos 8 caracteres")
	ErrPasswordIgual            = errors.New("la nueva contraseña debe ser diferente a la actual")
	ErrFormatoImagenInvalido    = errors.New("formato de imagen no permitido (solo PNG, JPG o WEBP)")
	ErrImagenDemasiadoGrande    = errors.New("el archivo excede el tamaño máximo de 5MB")
)

const MaxImageSizeBytes = 5 * 1024 * 1024 // 5MB

func (s *MiPerfilService) ObtenerPerfil(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error) {
	if uuidPersona.IsNil() {
		return miperfil.PerfilUsuario{}, errors.New("uuid de persona inválido")
	}
	return s.repo.ObtenerPerfilUsuario(ctx, uuidPersona)
}

func (s *MiPerfilService) CambiarPassword(ctx context.Context, uuidPersona uuid.UUID, form miperfil.CambioPasswordForm) error {
	if uuidPersona.IsNil() {
		return errors.New("uuid de persona inválido")
	}

	form.PasswordActual = strings.TrimSpace(form.PasswordActual)
	form.PasswordNueva = strings.TrimSpace(form.PasswordNueva)
	form.PasswordConfirmacion = strings.TrimSpace(form.PasswordConfirmacion)

	if form.PasswordActual == "" {
		return errors.New("la contraseña actual es requerida")
	}
	if len(form.PasswordNueva) < 8 {
		return ErrPasswordCorta
	}
	if form.PasswordNueva != form.PasswordConfirmacion {
		return ErrPasswordNoCoincide
	}
	if form.PasswordActual == form.PasswordNueva {
		return ErrPasswordIgual
	}

	hashActual, err := s.repo.ObtenerPasswordHash(ctx, uuidPersona)
	if err != nil {
		return fmt.Errorf("error al verificar credenciales actuales: %w", err)
	}

	if !password.CheckPasswordHash(form.PasswordActual, hashActual) {
		return ErrPasswordActualIncorrecta
	}

	nuevoHash, err := password.HashPassword(form.PasswordNueva)
	if err != nil {
		return fmt.Errorf("error al hashear nueva contraseña: %w", err)
	}

	return s.repo.ActualizarPassword(ctx, uuidPersona, nuevoHash)
}

func (s *MiPerfilService) CambiarImagen(ctx context.Context, uuidPersona uuid.UUID, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	if uuidPersona.IsNil() {
		return "", errors.New("uuid de persona inválido")
	}
	if size > MaxImageSizeBytes {
		return "", ErrImagenDemasiadoGrande
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		// Extensión permitida
	default:
		return "", ErrFormatoImagenInvalido
	}

	// Normalizar extensión
	if ext == ".jpeg" {
		ext = ".jpg"
	}

	// Leer buffer inicial para validar magic bytes
	headerBytes := make([]byte, 512)
	n, err := io.ReadFull(reader, headerBytes)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("error leyendo contenido de imagen: %w", err)
	}
	headerBytes = headerBytes[:n]

	// Reensamblar reader completo
	fullReader := io.MultiReader(bytes.NewReader(headerBytes), reader)

	nombreArchivo := fmt.Sprintf("%s_%d%s", uuidPersona.String(), time.Now().Unix(), ext)
	rutaRelativa := filepath.Join("img", "perfil", nombreArchivo)

	// Guardar mediante ManejadorArchivos si está configurado
	if s.archivos != nil {
		archivoPublico := s.archivos.ArchivoPublico(rutaRelativa)
		if err := archivoPublico.Guardar(ctx, fullReader); err != nil {
			return "", fmt.Errorf("error guardando archivo con manejador: %w", err)
		}
	} else {
		// Fallback directo a static/public/img/perfil
		destDir := filepath.Join("static", "public", "img", "perfil")
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return "", fmt.Errorf("error creando directorio destino: %w", err)
		}
		destPath := filepath.Join(destDir, nombreArchivo)
		f, err := os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("error creando archivo en disco: %w", err)
		}
		defer f.Close()
		if _, err := io.Copy(f, fullReader); err != nil {
			return "", fmt.Errorf("error escribiendo imagen en disco: %w", err)
		}
	}

	// Persistir en base de datos
	if err := s.repo.ActualizarImagenPerfil(ctx, uuidPersona, nombreArchivo); err != nil {
		return "", fmt.Errorf("error actualizando registro de imagen: %w", err)
	}

	return nombreArchivo, nil
}
