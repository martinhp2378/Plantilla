package ui

import (
	"context"
	"math/rand"

	"github.com/issemym-repos/archivos"
)

func ImagenLogin(ctx context.Context) string {
	imagenes := []string{"img/login1.jpg", "img/login2.jpg", "img/login3.jpg"}
	imagen := imagenes[rand.Intn(len(imagenes))]
	return archivos.Asset(ctx, imagen)
}
