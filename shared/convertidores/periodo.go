package convertidores

import (
	"PlantillaInstitucional/shared/views/components"
)

func OptionFromPeriodo() []components.SelectOption {
	return []components.SelectOption{
		{Value: "", Text: "Seleccione un periodo"},
		{Value: "ENERO-JUNIO", Text: "ENERO-JUNIO"},
		{Value: "JULIO-DICIEMBRE", Text: "JULIO-DICIEMBRE"},
	}
}
