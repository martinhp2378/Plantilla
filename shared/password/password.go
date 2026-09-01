package password

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordChars      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@#$%&()-_?"
	passwordLength     = 12
	usernamePrefix     = "user_"
	usernameChars      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	usernameRandLength = 8
)

// GenerateRandomPassword genera una contraseña segura
func GenerateRandomPassword() (string, error) {
	return generateRandomString(passwordChars, passwordLength)
}

// GenerateOpaqueToken returns a URL-safe, high-entropy token suitable for a
// one-time activation or password-reset link.
func GenerateOpaqueToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// HashOpaqueToken returns the only representation of an activation token that
// may be persisted.
func HashOpaqueToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// GenerateRandomUsername genera un nombre de usuario aleatorio
func GenerateRandomUsername() (string, error) {
	randomPart, err := generateRandomString(usernameChars, usernameRandLength)
	if err != nil {
		return "", err
	}
	return usernamePrefix + randomPart, nil
}

// Función genérica para generar strings aleatorios
func generateRandomString(charset string, length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando bytes aleatorios: %w", err)
	}

	for i := 0; i < length; i++ {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(bytes), nil
}

// HashPassword hashea una contraseña usando bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hasheando contraseña: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash compara una contraseña con su hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
