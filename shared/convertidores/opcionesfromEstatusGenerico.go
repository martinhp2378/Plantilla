package convertidores

import (
	"PlantillaInstitucional/shared/views/components"
)

var opciones = []components.SelectOption{
	{Value: "1", Text: "Activo"},
	{Value: "-1", Text: "Inactivo"},
}

func OptionFromEstatusGenerico() []components.SelectOption {
	options := make([]components.SelectOption, 0, 1+len(opciones))
	options = append(options, components.SelectOption{
		Value: "",
		Text:  "Seleccione un Estatus",
	})

	for _, v := range opciones {
		options = append(options, components.SelectOption{
			Value: v.Value,
			Text:  v.Text,
		})
	}
	return options
}

func Traducir(b bool) string {
	if b {
		return "1"
	}
	return "-1"
}
