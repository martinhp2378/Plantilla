package convertidores

import (
	"PlantillaInstitucional/shared/views/components"
)

func OpcionFromTipoUnidad() []components.SelectOption {
	var options []components.SelectOption
	options = append(options, components.SelectOption{
		Value: "",
		Text:  "Seleccione un Tipo de Unidad",
	})
	options = append(options, components.SelectOption{
		Value: "ADMINISTRATIVA",
		Text:  "Administrativa",
	})
	options = append(options, components.SelectOption{
		Value: "MEDICA",
		Text:  "Médica",
	})
	return options
}

func OpcionFromNivel() []components.SelectOption {
	var options []components.SelectOption
	options = append(options, components.SelectOption{
		Value: "",
		Text:  "Seleccione un Nivel",
	})
	options = append(options, components.SelectOption{
		Value: "1",
		Text:  "Nivel 1",
	})
	options = append(options, components.SelectOption{
		Value: "2",
		Text:  "Nivel 2",
	})
	options = append(options, components.SelectOption{
		Value: "3",
		Text:  "Nivel 3",
	})
	options = append(options, components.SelectOption{
		Value: "99",
		Text:  "Nivel 99 (Administrativa)",
	})
	return options
}