package email

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

const (
	// DefaultTimeout is the default timeout for email operations
	DefaultTimeout = 30 * time.Second
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

type ParameterProvider interface {
	GetParameter(key string) string
}

type TemplateBranding struct {
	SystemName          string
	BrandName           string
	FooterOrganization  string
	FooterLocation      string
	SupportLabel        string
	SupportPhone        string
	NoReplyLabel        string
	CredentialsEyebrow  string
	CredentialsIntro    string
	CredentialsSecurity string
	CredentialsURL      string
	CredentialsButton   string
}

type UserCredentialsTemplateData struct {
	TemplateBranding
	Nombre     string
	Usuario    string
	Password   string
	AnioActual int
}

// EmailService handles email sending logic.
type EmailService struct {
	config     Config
	parameters ParameterProvider
}

// NewEmailService creates a new instance of EmailService.
func NewEmailService(config Config, parameterProviders ...ParameterProvider) *EmailService {
	var provider ParameterProvider
	if len(parameterProviders) > 0 {
		provider = parameterProviders[0]
	}
	return &EmailService{
		config:     config,
		parameters: provider,
	}
}

// SetParametersProvider sets the provider for branding parameters.
func (s *EmailService) SetParametersProvider(provider ParameterProvider) {
	s.parameters = provider
}

func (s *EmailService) templateBranding() TemplateBranding {
	return TemplateBranding{
		SystemName:          s.param("EMAIL_SYSTEM_NAME", "Plantilla Institucional"),
		BrandName:           s.param("EMAIL_BRAND_NAME", "ISSEMYM"),
		FooterOrganization:  s.param("EMAIL_FOOTER_ORGANIZATION", "ISSEMYM - Unidad de Tecnologías"),
		FooterLocation:      s.param("EMAIL_FOOTER_LOCATION", "Toluca, Estado de México."),
		SupportLabel:        s.param("EMAIL_SUPPORT_LABEL", "¿Necesitas ayuda?"),
		SupportPhone:        s.param("EMAIL_SUPPORT_PHONE", "722 226 1900 ext. 1911"),
		NoReplyLabel:        s.param("EMAIL_NO_REPLY_LABEL", "MENSAJE AUTOMÁTICO - NO RESPONDER"),
		CredentialsEyebrow:  s.param("EMAIL_CREDENTIALS_EYEBROW", "Acceso al Sistema"),
		CredentialsIntro:    s.param("EMAIL_CREDENTIALS_INTRO", "Tu cuenta ha sido habilitada en la Plantilla Institucional. A continuación se presentan tus credenciales de acceso:"),
		CredentialsSecurity: s.param("EMAIL_CREDENTIALS_SECURITY", "Por tu seguridad, el sistema te solicitará cambiar esta contraseña en tu primer inicio de sesión."),
		CredentialsURL:      s.param("EMAIL_CREDENTIALS_URL", s.param("URL_BASE", "http://localhost:8080/")),
		CredentialsButton:   s.param("EMAIL_CREDENTIALS_BUTTON", "Acceder al Sistema"),
	}
}

func (s *EmailService) param(key, fallback string) string {
	if s.parameters == nil {
		return fallback
	}
	value := strings.TrimSpace(s.parameters.GetParameter(key))
	if value == "" {
		return fallback
	}
	return value
}

// SendUserCredentials sends an email with the user's credentials.
func (s *EmailService) SendUserCredentials(to, name, username, password string, credentialsURL ...string) error {
	branding := s.templateBranding()
	if len(credentialsURL) > 0 && strings.TrimSpace(credentialsURL[0]) != "" {
		branding.CredentialsURL = strings.TrimSpace(credentialsURL[0])
	}

	data := UserCredentialsTemplateData{
		TemplateBranding: branding,
		Nombre:           name,
		Usuario:          username,
		Password:         password,
		AnioActual:       time.Now().Year(),
	}

	body, err := s.ParseTemplate("user_credentials.html", data)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()

	from := s.config.From
	if s.config.FromName != "" {
		from = m.FormatAddress(s.config.From, s.config.FromName)
	}
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Bienvenido a la Plantilla Institucional - Credenciales de Acceso")
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.Host, 587, s.config.User, s.config.Password)

	errChan := make(chan error, 1)
	go func() {
		errChan <- d.DialAndSend(m)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(DefaultTimeout):
		return fmt.Errorf("timeout al enviar correo después de %v", DefaultTimeout)
	}
}

// ParseTemplate parses an HTML template with the provided data.
func (s *EmailService) ParseTemplate(templateName string, data interface{}) (string, error) {
	tmplPath := filepath.Join("shared", "email", "templates", templateName)

	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template %s: %w", templateName, err)
	}

	return buf.String(), nil
}
