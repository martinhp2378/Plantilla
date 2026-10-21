package cargamasiva

import (
	"strings"
	"testing"
)

type PersonaPrueba struct {
	Nombre   string
	Edad     int
	Sueldo   float64
	Activo   bool
	Fecha    string
	Correo   string
	TipoUser string
}

func configPrueba() ConfigCarga[PersonaPrueba] {
	return ConfigCarga[PersonaPrueba]{
		ColumnasEsperadas: []DefinicionColumna[PersonaPrueba]{
			{
				Nombre:    "nombre",
				Requerido: true,
				Ejemplo:   "Juan Pérez",
				Parser: func(valor string, e *PersonaPrueba) error {
					e.Nombre = valor
					return nil
				},
			},
			{
				Nombre:    "edad",
				Requerido: true,
				Ejemplo:   "30",
				Parser: func(valor string, e *PersonaPrueba) error {
					n, err := ParseEntero(valor, 18, 100)
					if err != nil {
						return err
					}
					e.Edad = n
					return nil
				},
			},
			{
				Nombre:    "sueldo",
				Requerido: false,
				Ejemplo:   "1500.50",
				Parser: func(valor string, e *PersonaPrueba) error {
					s, err := ParseDecimal(valor, 0, 1000000)
					if err != nil {
						return err
					}
					e.Sueldo = s
					return nil
				},
			},
			{
				Nombre:    "activo",
				Requerido: true,
				Ejemplo:   "Sí",
				Parser: func(valor string, e *PersonaPrueba) error {
					b, err := ParseBooleano(valor)
					if err != nil {
						return err
					}
					e.Activo = b
					return nil
				},
			},
			{
				Nombre:    "fecha",
				Requerido: true,
				Ejemplo:   "2026-10-21",
				Parser: func(valor string, e *PersonaPrueba) error {
					f, err := ParseFecha(valor)
					if err != nil {
						return err
					}
					e.Fecha = f
					return nil
				},
			},
			{
				Nombre:    "correo",
				Requerido: true,
				Ejemplo:   "juan@ejemplo.com",
				Parser: func(valor string, e *PersonaPrueba) error {
					c, err := ParseEmail(valor)
					if err != nil {
						return err
					}
					e.Correo = c
					return nil
				},
			},
			{
				Nombre:    "tipo",
				Requerido: false,
				Ejemplo:   "ADMIN",
				Parser: func(valor string, e *PersonaPrueba) error {
					t, err := ParseEnum(valor, []string{"ADMIN", "USER"})
					if err != nil {
						return err
					}
					e.TipoUser = t
					return nil
				},
			},
		},
	}
}

func TestProcesarCSV_Exitoso(t *testing.T) {
	csvData := "\xef\xbb\xbfnombre,EDAD,sueldo,activo,fecha,correo,tipo\n" +
		"Ana Gómez,25,2500.75,Sí,2026-05-15,ana@issemym.gob.mx,ADMIN\n" +
		"Carlos Ruiz,40,1800.00,no,2026-08-20,carlos@issemym.gob.mx,USER\n"

	res, err := ProcesarCSV(strings.NewReader(csvData), configPrueba())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if !res.EsExitoso() {
		t.Fatalf("se esperaba carga exitosa, errores: %v", res.Errores)
	}

	if res.TotalValidos != 2 {
		t.Errorf("se esperaban 2 válidos, se obtuvieron %d", res.TotalValidos)
	}

	if res.Registros[0].Nombre != "Ana Gómez" || res.Registros[0].Edad != 25 || !res.Registros[0].Activo {
		t.Errorf("primer registro no coincide: %+v", res.Registros[0])
	}
	if res.Registros[1].Nombre != "Carlos Ruiz" || res.Registros[1].Activo != false {
		t.Errorf("segundo registro no coincide: %+v", res.Registros[1])
	}
}

func TestProcesarCSV_FaltaColumnaObligatoria(t *testing.T) {
	csvData := "nombre,edad,activo\nAna,25,Sí\n"

	_, err := ProcesarCSV(strings.NewReader(csvData), configPrueba())
	if err == nil {
		t.Fatal("se esperaba error por falta de columnas obligatorias")
	}
	if !strings.Contains(err.Error(), "faltan columnas obligatorias") {
		t.Errorf("mensaje de error inesperado: %v", err)
	}
}

func TestProcesarCSV_ErroresDeValidacionPorFila(t *testing.T) {
	csvData := "nombre,edad,sueldo,activo,fecha,correo,tipo\n" +
		",no_es_numero,abc,invalido,2026/10/21,correo_mal,DESCONOCIDO\n"

	res, err := ProcesarCSV(strings.NewReader(csvData), configPrueba())
	if err != nil {
		t.Fatalf("error inesperado procesando CSV: %v", err)
	}

	if res.TotalErrores != 1 {
		t.Errorf("se esperaba 1 fila con errores, se obtuvieron %d", res.TotalErrores)
	}

	if len(res.Errores) < 6 {
		t.Errorf("se esperaban al menos 6 errores detallados de columna, se obtuvieron %d: %+v", len(res.Errores), res.Errores)
	}
}

func TestGenerarPlantillaCSV(t *testing.T) {
	cfg := configPrueba()
	bytesCSV, err := GenerarPlantillaCSV(cfg)
	if err != nil {
		t.Fatalf("error al generar plantilla: %v", err)
	}

	contenido := string(bytesCSV)
	if !strings.Contains(contenido, "nombre,edad,sueldo,activo,fecha,correo,tipo") {
		t.Errorf("encabezado de plantilla incorrecto: %s", contenido)
	}
	if !strings.Contains(contenido, "Juan Pérez,30,1500.50,Sí,2026-10-21,juan@ejemplo.com,ADMIN") {
		t.Errorf("fila de ejemplo incorrecta: %s", contenido)
	}
}
