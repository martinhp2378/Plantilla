package db

import (
	"context"
	"errors"
	"testing"

	contextkey "PlantillaInstitucional/shared/contextkey"

	"github.com/jackc/pgx/v5/pgconn"
)

// mockExecutor es un Executor simulado para testing
type mockExecutor struct {
	execCalls   []mockExecCall
	errToReturn error
}

type mockExecCall struct {
	ctx       context.Context
	sql       string
	arguments []any
}

func (m *mockExecutor) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	m.execCalls = append(m.execCalls, mockExecCall{
		ctx:       ctx,
		sql:       sql,
		arguments: arguments,
	})
	if m.errToReturn != nil {
		return pgconn.NewCommandTag(""), m.errToReturn
	}
	return pgconn.NewCommandTag("SET"), nil
}

// TestSetAuditSession_WithValidUUID verifica que SetAuditSession envíe el UUID correcto
func TestSetAuditSession_WithValidUUID(t *testing.T) {
	testUUID := "550e8400-e29b-41d4-a716-446655440000"
	ctx := context.WithValue(context.Background(), contextkey.AuditUUIDKey, testUUID)
	mock := &mockExecutor{}

	err := SetAuditSession(ctx, mock)

	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

	if len(mock.execCalls) != 1 {
		t.Fatalf("Se esperaba 1 llamada a Exec(), se hicieron %d", len(mock.execCalls))
	}

	call := mock.execCalls[0]
	expectedSQL := "SELECT set_config('session_variables.uuid_persona', $1, false)"
	if call.sql != expectedSQL {
		t.Errorf("SQL incorrecto. Esperado: %q, obtuvo: %q", expectedSQL, call.sql)
	}

	if call.arguments[0] != testUUID {
		t.Errorf("UUID incorrecto. Esperado: %q, obtuvo: %q", testUUID, call.arguments[0])
	}
}

// TestSetAuditSession_WithoutUUIDInContext verifica el comportamiento sin UUID
func TestSetAuditSession_WithoutUUIDInContext(t *testing.T) {
	ctx := context.Background()
	mock := &mockExecutor{}

	err := SetAuditSession(ctx, mock)

	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

	call := mock.execCalls[0]
	if call.arguments[0] != "" {
		t.Errorf("Debería enviar cadena vacía sin UUID, obtuvo: %q", call.arguments[0])
	}
}

// TestSetAuditSession_WrongType verifica que tipos incorrectos resulten en cadena vacía
func TestSetAuditSession_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkey.AuditUUIDKey, 12345)
	mock := &mockExecutor{}

	err := SetAuditSession(ctx, mock)

	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

	call := mock.execCalls[0]
	if call.arguments[0] != "" {
		t.Errorf("Tipo incorrecto debería resultar en cadena vacía, obtuvo: %q", call.arguments[0])
	}
}

// TestSetAuditSession_ExecutorError verifica propagación de errores
func TestSetAuditSession_ExecutorError(t *testing.T) {
	testUUID := "550e8400-e29b-41d4-a716-446655440000"
	ctx := context.WithValue(context.Background(), contextkey.AuditUUIDKey, testUUID)
	expectedErr := errors.New("conexión perdida")
	mock := &mockExecutor{errToReturn: expectedErr}

	err := SetAuditSession(ctx, mock)

	if err == nil {
		t.Fatal("Se esperaba un error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Error incorrecto. Esperado: %v, obtuvo: %v", expectedErr, err)
	}
}

// TestSetAuditSession_TableDriven prueba múltiples UUIDs con tabla de casos
func TestSetAuditSession_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		contextValue any
		expectedArg  string
	}{
		{"UUID válido", "f47ac10b-58cc-4372-a567-0e02b2c3d479", "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
		{"Sin UUID", nil, ""},
		{"Tipo entero", 999, ""},
		{"Booleano", false, ""},
		{"Cadena vacía", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.contextValue != nil {
				ctx = context.WithValue(ctx, contextkey.AuditUUIDKey, tt.contextValue)
			}
			mock := &mockExecutor{}

			err := SetAuditSession(ctx, mock)

			if err != nil {
				t.Fatalf("Error inesperado: %v", err)
			}

			call := mock.execCalls[0]
			if call.arguments[0] != tt.expectedArg {
				t.Errorf("Argumento incorrecto. Esperado: %q, obtuvo: %q", tt.expectedArg, call.arguments[0])
			}
		})
	}
}

// BenchmarkSetAuditSession mide el rendimiento de SetAuditSession
func BenchmarkSetAuditSession(b *testing.B) {
	testUUID := "550e8400-e29b-41d4-a716-446655440000"
	ctx := context.WithValue(context.Background(), contextkey.AuditUUIDKey, testUUID)
	mock := &mockExecutor{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Resetear calls para no acumular
		mock.execCalls = nil
		_ = SetAuditSession(ctx, mock)
	}
}
