package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// GenerateQR crea una imagen QR en formato PNG a partir de un contenido string.
// size define el ancho y alto en píxeles (ej. 256).
// Retorna los bytes de la imagen PNG o un error.
func GenerateQR(content string, size int) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("el contenido del QR no puede estar vacío")
	}

	if size <= 0 {
		size = 256 // Valor por defecto
	}

	// Crear el código QR
	qrCode, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("error codificando QR: %w", err)
	}

	// Escalar el código QR al tamaño deseado
	qrCode, err = barcode.Scale(qrCode, size, size)
	if err != nil {
		return nil, fmt.Errorf("error escalando QR: %w", err)
	}

	// Convertir a RGBA explícito para evitar problemas de profundidad de bits (ej. 16-bit grayscale) con librerías de PDF
	rect := qrCode.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, qrCode, rect.Min, draw.Src)

	// Convertir a PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, fmt.Errorf("error codificando a PNG: %w", err)
	}

	return buf.Bytes(), nil
}
