package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakeRecoveryStore struct {
	found  bool
	issued bool
}

func (s *fakeRecoveryStore) IssuePasswordReset(_ context.Context, _ string, tokenHash []byte, _ time.Time) (bool, error) {
	s.issued = len(tokenHash) == 32
	return s.found, nil
}

func (s *fakeRecoveryStore) ConsumePasswordReset(context.Context, []byte, string) error { return nil }

type fakeActivationSender struct{ sent bool }

func (s *fakeActivationSender) SendUserActivation(context.Context, string, string, string, string) error {
	s.sent = true
	return nil
}

func TestPasswordRecoveryDoesNotEnumerateAccounts(t *testing.T) {
	for _, found := range []bool{false, true} {
		t.Run(map[bool]string{false: "unknown account", true: "known account"}[found], func(t *testing.T) {
			store := &fakeRecoveryStore{found: found}
			sender := &fakeActivationSender{}
			h := &HandlerLogin{AccountRecoveryProvider: store, activationEmailSender: sender, resetLimiter: newResetRateLimiter()}
			req := httptest.NewRequest(http.MethodPost, "/login/recuperar", strings.NewReader(url.Values{"email": {"person@example.test"}}.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			h.HandlePasswordRecovery(rec, req)
			if rec.Code != http.StatusAccepted || rec.Body.String() != "Si existe una cuenta para este correo, recibirás instrucciones para continuar." {
				t.Fatalf("response = %d %q, want non-enumerating accepted response", rec.Code, rec.Body.String())
			}
			if !store.issued {
				t.Fatal("recovery token must be hash-sized before persistence")
			}
			if sender.sent != found {
				t.Fatalf("sender.sent = %t, want %t", sender.sent, found)
			}
		})
	}
}

func TestPasswordRecoveryRateLimit(t *testing.T) {
	limiter := newResetRateLimiter()
	limiter.now = func() time.Time { return time.Unix(0, 0) }
	for range resetMaxTries {
		if !limiter.allow("127.0.0.1") {
			t.Fatal("request inside limit was rejected")
		}
	}
	if limiter.allow("127.0.0.1") {
		t.Fatal("request above limit was allowed")
	}
}
