package elver

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Elver struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Nombre      string    `db:"nombre" json:"nombre" validate:"required,min=2,max=150"`
	Codigo      string    `db:"codigo" json:"codigo" validate:"required,min=2,max=50"`
	Descripcion string    `db:"descripcion" json:"descripcion"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	FechaI      time.Time `db:"fecha_i" json:"fecha_i"`
	FechaM      time.Time `db:"fecha_m" json:"fecha_m"`
}
