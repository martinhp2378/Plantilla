package bd

import (
	"PlantillaInstitucional/shared/entorno"
	"testing"
)

func TestBuildConnectionURL(t *testing.T) {
	t.Run("Standard URL", func(t *testing.T) {
		env := &entorno.Env{
			PqWriteUser:    "user",
			PqWritePass:    "pass",
			PqWriteHost:    "host",
			PqWritePort:    "5432",
			PqWriteDb:      "db",
			PqWriteSSLMode: "disable",
		}

		url := BuildConnectionURL(env, true)
		expected := "postgres://user:pass@host:5432/db?sslmode=disable"
		if url != expected {
			t.Errorf("expected %s, got %s", expected, url)
		}
	})

	t.Run("CloudSQL URL", func(t *testing.T) {
		env := &entorno.Env{
			PqWriteUser:    "user",
			PqWritePass:    "pass",
			PqWriteHost:    "/cloudsql/project:region:instance",
			PqWriteDb:      "db",
			PqWriteSSLMode: "disable",
		}

		url := BuildConnectionURL(env, true)
		// postgres://user:pass@localhost/db?host=%2Fcloudsql%2Fproject%3Aregion%3Ainstance&sslmode=disable
		if !contains(url, "host=%2Fcloudsql%2F") {
			t.Errorf("expected CloudSQL URL, got %s", url)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))
}
