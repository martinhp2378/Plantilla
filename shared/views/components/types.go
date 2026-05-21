package components

type CardVista struct {
	Title       string
	Value       string
	Icon        string
	IconClass   string
	BgClass     string
	TextClass   string
	BorderClass string
	Meta        string
}

type SelectOption struct {
	Value string
	Text  string
}

type GraficaDimensionesAvg struct {
	DimensionID string  `db:"dimension_id" json:"dimension_id"`
	Nombre      string  `db:"nombre_dimension" json:"nombre_dimension"`
	Promedio    float64 `db:"promedio_obtenido" json:"promedio"`
}

type GraficaIndicadoresAvg struct {
	IndicadorID     string  `db:"IndicadorID" json:"indicador_id"`
	IndicadorNombre string  `db:"IndicadorNombre" json:"indicador_nombre"`
	Promedio        float64 `db:"Promedio" json:"promedio"`
}

type GraficaBarrasMultiples struct {
	Dimension string `db:"dimension" json:"dimension"`
	Respuesta int    `db:"respuesta" json:"respuesta"`
	Total     int    `db:"total" json:"total"`
}

type GraficaPie struct {
	Respuesta int `db:"respuesta" json:"respuesta"`
	Total     int `db:"total" json:"total"`
}

type ScriptsDashboard struct {
	TieneLine              bool
	JsonLine               string
	TieneRadar             bool
	JsonRadar              string
	TieneMultiBar          bool
	JsonMultiBar           string
	TienePie               bool
	JsonPie                string
	DimensionesRealData    string
	DimensionesObjetivoData string
}
