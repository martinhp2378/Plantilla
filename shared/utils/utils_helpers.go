package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SafeStringPtr returns the string value of the pointer or an empty string if nil
func SafeStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// FormatDatePtr returns the formatted date string or an empty string if nil
func FormatDatePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// GetBaseURL resuelve de forma dinámica el protocolo y host (incluyendo subdominio)
// para construir URLs absolutas correctas en entornos multi-tenant.
func GetBaseURL(r *http.Request) string {
	baseURL, err := GetTrustedBaseURL(r)
	if err == nil {
		return baseURL
	}

	trustProxy := strings.EqualFold(strings.TrimSpace(os.Getenv("TRUST_PROXY_HEADERS")), "true")

	scheme := "http"
	// Detectar si es HTTPS (directo o via proxy)
	if r.TLS != nil || (trustProxy && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")) {
		scheme = "https"
	}

	// Obtener el Host (r.Host ya incluye el puerto si es necesario)
	host := r.Host

	// Si estamos detrás de un proxy bien configurado, X-Forwarded-Host es más confiable
	if trustProxy {
		if fwdHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); fwdHost != "" {
			host = strings.Split(fwdHost, ",")[0]
		}
	}

	host = strings.TrimSpace(host)

	return scheme + "://" + host
}

// GetTrustedBaseURL construye una URL base segura para enlaces externos.
// Prioriza APP_BASE_URL, luego host/proxy sanitizados con allowlist opcional.
func GetTrustedBaseURL(r *http.Request) (string, error) {
	if raw := strings.TrimSpace(os.Getenv("APP_BASE_URL")); raw != "" {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "", fmt.Errorf("APP_BASE_URL inválida")
		}
		return strings.TrimRight(raw, "/"), nil
	}

	host := strings.TrimSpace(r.Host)
	trustProxy := strings.EqualFold(strings.TrimSpace(os.Getenv("TRUST_PROXY_HEADERS")), "true")
	if trustProxy {
		if fwdHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); fwdHost != "" {
			host = strings.Split(fwdHost, ",")[0]
		}
	}

	host = strings.TrimSpace(host)
	if strings.ContainsAny(host, "/\\ ") || strings.Contains(host, ",") || host == "" {
		return "", fmt.Errorf("host inválido")
	}

	allowed := parseAllowedHosts(os.Getenv("TRUSTED_BASE_HOSTS"))
	if len(allowed) > 0 && !allowed[strings.ToLower(host)] {
		return "", fmt.Errorf("host no permitido")
	}

	scheme := "http"
	if r.TLS != nil || (trustProxy && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")) {
		scheme = "https"
	}

	return scheme + "://" + host, nil
}

func parseAllowedHosts(raw string) map[string]bool {
	allowed := make(map[string]bool)
	for _, p := range strings.Split(raw, ",") {
		h := strings.ToLower(strings.TrimSpace(p))
		if h != "" {
			allowed[h] = true
		}
	}
	return allowed
}
