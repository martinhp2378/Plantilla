package db

import (
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestParseUUID(t *testing.T) {
	valid := "550e8400-e29b-41d4-a716-446655440000"
	parsedValid, err := uuid.FromString(valid)
	if err != nil {
		t.Fatalf("setup failed parsing valid uuid: %v", err)
	}

	tests := []struct {
		name       string
		input      string
		want       uuid.UUID
		wantErr    bool
		errMessage string
	}{
		{name: "valid uuid", input: valid, want: parsedValid, wantErr: false},
		{name: "empty uuid returns nil", input: "", want: uuid.Nil, wantErr: false},
		{name: "invalid uuid format", input: "not-a-uuid", want: uuid.Nil, wantErr: true, errMessage: "ID con formato inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUUID(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("uuid = %v, want %v", got, tt.want)
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Fatalf("error message = %q, want %q", err.Error(), tt.errMessage)
			}
		})
	}
}

func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		def    int
		expect int
	}{
		{name: "positive number", input: "12", def: 5, expect: 12},
		{name: "zero returns default", input: "0", def: 5, expect: 5},
		{name: "negative returns default", input: "-3", def: 5, expect: 5},
		{name: "invalid string returns default", input: "x", def: 5, expect: 5},
		{name: "empty string returns default", input: "", def: 8, expect: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIntOrDefault(tt.input, tt.def)
			if got != tt.expect {
				t.Fatalf("value = %d, want %d", got, tt.expect)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       time.Time
		wantZero   bool
		wantErr    bool
		errMessage string
	}{
		{
			name:     "valid date",
			input:    "2024-03-25",
			want:     time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC),
			wantZero: false,
			wantErr:  false,
		},
		{name: "empty returns zero", input: "", wantZero: true, wantErr: false},
		{name: "invalid format", input: "25-03-2024", wantZero: true, wantErr: true, errMessage: "formato de fecha inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantZero {
				if !got.IsZero() {
					t.Fatalf("time = %v, want zero time", got)
				}
				return
			}

			if !got.Equal(tt.want) {
				t.Fatalf("time = %v, want %v", got, tt.want)
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMessage) {
				t.Fatalf("error message = %q, expected to contain %q", err.Error(), tt.errMessage)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       int
		wantErr    bool
		errMessage string
	}{
		{name: "valid int", input: "42", want: 42, wantErr: false},
		{name: "empty returns zero", input: "", want: 0, wantErr: false},
		{name: "invalid int", input: "3.14", want: 0, wantErr: true, errMessage: "valor de número inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("value = %d, want %d", got, tt.want)
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMessage) {
				t.Fatalf("error message = %q, expected to contain %q", err.Error(), tt.errMessage)
			}
		})
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       float64
		wantErr    bool
		errMessage string
	}{
		{name: "valid float", input: "3.14", want: 3.14, wantErr: false},
		{name: "empty returns zero", input: "", want: 0.0, wantErr: false},
		{name: "invalid float", input: "abc", want: 0.0, wantErr: true, errMessage: "valor de número flotante inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloat64(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("value = %v, want %v", got, tt.want)
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMessage) {
				t.Fatalf("error message = %q, expected to contain %q", err.Error(), tt.errMessage)
			}
		})
	}
}
