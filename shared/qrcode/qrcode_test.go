package qrcode

import (
	"bytes"
	"image/png"
	"strings"
	"testing"
)

func TestGenerateQRReturnsPNGWithExpectedDimensions(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		size       int
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "uses requested size",
			content:    "https://example.com/entidad/folio-123",
			size:       128,
			wantWidth:  128,
			wantHeight: 128,
		},
		{
			name:       "uses default size when size is zero",
			content:    "folio-456",
			size:       0,
			wantWidth:  256,
			wantHeight: 256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := GenerateQR(tt.content, tt.size)
			if err != nil {
				t.Fatalf("GenerateQR() error = %v", err)
			}

			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("png.Decode() error = %v", err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.wantWidth || bounds.Dy() != tt.wantHeight {
				t.Fatalf("image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestGenerateQRReturnsErrorWhenContentIsEmpty(t *testing.T) {
	_, err := GenerateQR("", 256)
	if err == nil {
		t.Fatal("GenerateQR() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "no puede estar vacio") && !strings.Contains(err.Error(), "no puede estar vacío") {
		t.Fatalf("GenerateQR() error = %q, want content validation message", err.Error())
	}
}
