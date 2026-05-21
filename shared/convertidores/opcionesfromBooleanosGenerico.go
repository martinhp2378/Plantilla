package convertidores

import "PlantillaInstitucional/shared/views/components"

var opcionesBool = []components.SelectOption{
	{Value: "1", Text: "Si"},
	{Value: "-1", Text: "No"},
}

func OptionFromBooleanoGenerico() []components.SelectOption {
	var options []components.SelectOption
	options = append(options, components.SelectOption{
		Value: "",
		Text:  "Seleccione una Opcion",
	})

	for _, v := range opcionesBool {
		options = append(options, components.SelectOption{
			Value: v.Value,
			Text:  v.Text,
		})
	}
	return options
}
