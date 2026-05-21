package email

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fakeParameterProvider map[string]string

func (f fakeParameterProvider) GetParameter(key string) string {
	return f[key]
}

func chdirToModuleRoot(t *testing.T) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no se pudo resolver el archivo actual")
	}

	moduleRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("obtener directorio actual: %v", err)
	}

	if err := os.Chdir(moduleRoot); err != nil {
		t.Fatalf("cambiar al directorio del modulo: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(previousWD); err != nil {
			t.Fatalf("restaurar directorio actual: %v", err)
		}
	})
}

func TestNewEmailService(t *testing.T) {
	config := Config{
		Host:     "smtp.example.com",
		Port:     "587",
		User:     "usuario",
		Password: "secreto",
		From:     "noreply@example.com",
		FromName: "Plantilla",
		UseTLS:   true,
	}

	service := NewEmailService(config)

	if service == nil {
		t.Fatal("NewEmailService() devolvio nil")
	}

	if service.config != config {
		t.Fatalf("config = %#v, want %#v", service.config, config)
	}
}

func TestParseTemplate(t *testing.T) {
	chdirToModuleRoot(t)

	service := NewEmailService(Config{})
	currentYear := strconv.Itoa(time.Now().Year())

	tests := []struct {
		name         string
		templateName string
		data         any
		contains     []string
		notContains  []string
	}{
		{
			name:         "renderiza user credentials con datos del usuario",
			templateName: "user_credentials.html",
			data: UserCredentialsTemplateData{
				TemplateBranding: service.templateBranding(),
				Nombre:           "Ada Lovelace",
				Usuario:          "adal",
				Password:         "temporal123",
				AnioActual:       time.Now().Year(),
			},
			contains: []string{
				"Ada Lovelace",
				"adal",
				"temporal123",
				currentYear,
				"http://localhost:8080/",
			},
			notContains: []string{"{{.Nombre}}", "{{.Usuario}}", "{{.Password}}", "{{.AnioActual}}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := service.ParseTemplate(tt.templateName, tt.data)
			if err != nil {
				t.Fatalf("ParseTemplate(%q) error = %v", tt.templateName, err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(body, expected) {
					t.Fatalf("ParseTemplate(%q) no contiene %q", tt.templateName, expected)
				}
			}

			for _, unexpected := range tt.notContains {
				if strings.Contains(body, unexpected) {
					t.Fatalf("ParseTemplate(%q) aun contiene placeholder %q", tt.templateName, unexpected)
				}
			}
		})
	}
}

func TestTemplateBrandingUsesParameters(t *testing.T) {
	service := NewEmailService(Config{}, fakeParameterProvider{
		"EMAIL_BRAND_NAME":      "Marca Editable",
		"EMAIL_SUPPORT_PHONE":   "555 123 4567",
		"EMAIL_CREDENTIALS_URL": "https://sistema.issemym.gob.mx",
	})

	branding := service.templateBranding()

	if branding.BrandName != "Marca Editable" {
		t.Fatalf("BrandName = %q, want %q", branding.BrandName, "Marca Editable")
	}
	if branding.SupportPhone != "555 123 4567" {
		t.Fatalf("SupportPhone = %q, want %q", branding.SupportPhone, "555 123 4567")
	}
	if branding.CredentialsURL != "https://sistema.issemym.gob.mx" {
		t.Fatalf("CredentialsURL = %q, want custom URL", branding.CredentialsURL)
	}
}

func TestParseTemplateMissingTemplate(t *testing.T) {
	chdirToModuleRoot(t)

	service := NewEmailService(Config{})

	_, err := service.ParseTemplate("missing_template.html", nil)
	if err == nil {
		t.Fatal("ParseTemplate() esperaba error para template inexistente")
	}

	if !strings.Contains(err.Error(), "missing_template.html") {
		t.Fatalf("error = %q, esperaba nombre del template", err.Error())
	}

	if !strings.Contains(err.Error(), "error parsing template") {
		t.Fatalf("error = %q, esperaba prefijo de parseo", err.Error())
	}
}
