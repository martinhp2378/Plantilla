package utils

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetTrustedBaseURL(t *testing.T) {
	t.Run("uses APP_BASE_URL when configured", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "https://sistema.issemym.gob.mx")
		req := httptest.NewRequest("GET", "http://evil.local/path", nil)
		req.Host = "evil.local"

		got, err := GetTrustedBaseURL(req)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if got != "https://sistema.issemym.gob.mx" {
			t.Fatalf("base url = %q, want %q", got, "https://sistema.issemym.gob.mx")
		}
	})

	t.Run("rejects untrusted forwarded host", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("TRUST_PROXY_HEADERS", "true")
		t.Setenv("TRUSTED_BASE_HOSTS", "sistema.issemym.gob.mx")
		req := httptest.NewRequest("GET", "https://sistema.issemym.gob.mx/path", nil)
		req.Host = "sistema.issemym.gob.mx"
		req.Header.Set("X-Forwarded-Host", "evil.com")
		req.Header.Set("X-Forwarded-Proto", "https")

		_, err := GetTrustedBaseURL(req)
		if err == nil {
			t.Fatal("err = nil, want error for untrusted host")
		}
	})

	t.Run("uses trusted forwarded host", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("TRUST_PROXY_HEADERS", "true")
		t.Setenv("TRUSTED_BASE_HOSTS", "sistema.issemym.gob.mx")
		req := httptest.NewRequest("GET", "http://internal.service/path", nil)
		req.Host = "internal.service"
		req.Header.Set("X-Forwarded-Host", "sistema.issemym.gob.mx")
		req.Header.Set("X-Forwarded-Proto", "https")

		got, err := GetTrustedBaseURL(req)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if got != "https://sistema.issemym.gob.mx" {
			t.Fatalf("base url = %q, want %q", got, "https://sistema.issemym.gob.mx")
		}
	})

	t.Run("rejects malformed host values", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("TRUST_PROXY_HEADERS", "false")
		t.Setenv("TRUSTED_BASE_HOSTS", "")
		req := httptest.NewRequest("GET", "http://localhost/path", nil)
		req.Host = "good.com,bad.com"

		_, err := GetTrustedBaseURL(req)
		if err == nil {
			t.Fatal("err = nil, want invalid host error")
		}
	})

	t.Run("ignores forwarded proto when proxy headers are not trusted", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("TRUST_PROXY_HEADERS", "false")
		t.Setenv("TRUSTED_BASE_HOSTS", "sistema.issemym.gob.mx")
		req := httptest.NewRequest("GET", "http://sistema.issemym.gob.mx/path", nil)
		req.Host = "sistema.issemym.gob.mx"
		req.Header.Set("X-Forwarded-Proto", "https")

		got, err := GetTrustedBaseURL(req)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if got != "http://sistema.issemym.gob.mx" {
			t.Fatalf("base url = %q, want %q", got, "http://sistema.issemym.gob.mx")
		}
	})

	_ = os.Unsetenv("APP_BASE_URL")
}

func TestGetBaseURL_ProxyHeaderTrustGate(t *testing.T) {
	t.Run("ignores forwarded proto and host when proxy headers are not trusted", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "http://%%invalid")
		t.Setenv("TRUST_PROXY_HEADERS", "false")
		req := httptest.NewRequest("GET", "http://sistema.issemym.gob.mx/path", nil)
		req.Host = "sistema.issemym.gob.mx"
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "evil.example")

		got := GetBaseURL(req)
		if got != "http://sistema.issemym.gob.mx" {
			t.Fatalf("base url = %q, want %q", got, "http://sistema.issemym.gob.mx")
		}
	})

	t.Run("uses forwarded proto and host when proxy headers are trusted", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "http://%%invalid")
		t.Setenv("TRUST_PROXY_HEADERS", "true")
		req := httptest.NewRequest("GET", "http://sistema.issemym.gob.mx/path", nil)
		req.Host = "sistema.issemym.gob.mx"
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "proxy.sistema.issemym.gob.mx")

		got := GetBaseURL(req)
		if got != "https://proxy.sistema.issemym.gob.mx" {
			t.Fatalf("base url = %q, want %q", got, "https://proxy.sistema.issemym.gob.mx")
		}
	})
}
