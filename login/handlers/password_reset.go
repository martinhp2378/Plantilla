package handlers

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	loginDB "PlantillaInstitucional/login/db"
	"PlantillaInstitucional/shared/password"
	"PlantillaInstitucional/shared/utils"
)

const (
	resetWindow   = 15 * time.Minute
	resetMaxTries = 5
	tokenLifetime = 24 * time.Hour
)

type resetRateEntry struct {
	started time.Time
	count   int
}

type resetRateLimiter struct {
	mu      sync.Mutex
	entries map[string]resetRateEntry
	now     func() time.Time
}

func newResetRateLimiter() *resetRateLimiter {
	return &resetRateLimiter{entries: make(map[string]resetRateEntry), now: time.Now}
}

func (l *resetRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	entry := l.entries[key]
	if entry.started.IsZero() || now.Sub(entry.started) >= resetWindow {
		l.entries[key] = resetRateEntry{started: now, count: 1}
		return true
	}
	if entry.count >= resetMaxTries {
		return false
	}
	entry.count++
	l.entries[key] = entry
	return true
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (h *HandlerLogin) HandlePasswordRecovery(w http.ResponseWriter, r *http.Request) {
	if !h.resetLimiter.allow(remoteIP(r)) {
		http.Error(w, "Demasiadas solicitudes. Intenta más tarde.", http.StatusTooManyRequests)
		return
	}

	// This response deliberately does not disclose whether the email exists.
	email := strings.TrimSpace(r.PostFormValue("email"))
	token, err := password.GenerateOpaqueToken()
	if err == nil && h.AccountRecoveryProvider != nil && h.activationEmailSender != nil && email != "" {
		found, issueErr := h.IssuePasswordReset(r.Context(), email, password.HashOpaqueToken(token), time.Now().Add(tokenLifetime))
		if issueErr != nil {
			h.Logger.ErrorContext(r.Context(), "No se pudo solicitar recuperación", issueErr)
		} else if found {
			url := utils.GetBaseURL(r) + "/login/activar?token=" + token
			if sendErr := h.activationEmailSender.SendUserActivation(r.Context(), email, "", email, url); sendErr != nil {
				h.Logger.ErrorContext(r.Context(), "No se pudo enviar recuperación", sendErr)
			}
		}
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("Si existe una cuenta para este correo, recibirás instrucciones para continuar."))
}

func (h *HandlerLogin) HandleActivationForm(w http.ResponseWriter, r *http.Request) {
	token := html.EscapeString(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "Enlace de activación inválido.", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<form method="post" action="/login/activar"><input type="hidden" name="csrf_token" value="%s"><input type="hidden" name="token" value="%s"><label>Nueva contraseña <input type="password" name="password" required></label><button type="submit">Guardar contraseña</button></form>`, html.EscapeString(utils.GetCSRFToken(r.Context())), token)
}

func (h *HandlerLogin) HandleActivation(w http.ResponseWriter, r *http.Request) {
	token := r.PostFormValue("token")
	newPassword := r.PostFormValue("password")
	if token == "" || len(newPassword) < 12 || h.AccountRecoveryProvider == nil {
		http.Error(w, "El enlace o la contraseña no son válidos.", http.StatusBadRequest)
		return
	}
	hash, err := password.HashPassword(newPassword)
	if err == nil {
		err = h.ConsumePasswordReset(r.Context(), password.HashOpaqueToken(token), hash)
	}
	if err != nil {
		if err != loginDB.ErrInvalidOrExpiredResetToken {
			h.Logger.ErrorContext(r.Context(), "No se pudo activar la cuenta", err)
		}
		http.Error(w, "El enlace no es válido o expiró.", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/login/", http.StatusSeeOther)
}
