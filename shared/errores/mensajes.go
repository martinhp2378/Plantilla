package errores

const (
	// Mensajes generales
	MensajeSinRegistros     = "sin registros"
	MensajeEstructura       = "error en estructura"
	MensajeSinServicio      = "s/servicio"
	MensajeOperacionExitosa = "operación exitosa"
	MensajeOperacionError   = "error en la operación"
	MensajeRequerido        = "requerido"

	// Prefijos para mensajes requeridos
	Seleccionar = "favor de seleccionar: "
	Ingresar    = "favor de ingresar: "

	// Mensajes Operación Exitosa
	MensajeOperacionExitosaArchivoGenerado       = MensajeOperacionExitosa + " archivo generado: "
	MensajeOperacionExitosaConfiguracionGuardada = MensajeOperacionExitosa + " configuración guardada: "
	MensajeOperacionExitosaLogin                 = MensajeOperacionExitosa + " sesión iniciada: "
	MensajeOperacionExitosaLogout                = MensajeOperacionExitosa + " sesión cerrada: "
	MensajeOperacionExitosaPasswordActualizado   = MensajeOperacionExitosa + " contraseña actualizada: "
	MensajeOperacionExitosaRegistroActualizado   = MensajeOperacionExitosa + " registro actualizado: "
	MensajeOperacionExitosaRegistroCreado        = MensajeOperacionExitosa + " registro creado:"
	MensajeOperacionExitosaRegistroEliminado     = MensajeOperacionExitosa + " registro eliminado: "

	// Mensajes Operación Error Común
	MensajeOperacionErrorAccesoDenegado       = MensajeOperacionError + " acceso denegado: "
	MensajeOperacionErrorArchivoMuyGrande     = MensajeOperacionError + " el archivo es demasiado grande: "
	MensajeOperacionErrorArchivoNoEncontrado  = MensajeOperacionError + " archivo no encontrado: "
	MensajeOperacionErrorArchivoNoProcesado   = MensajeOperacionError + " archivo no procesado: "
	MensajeOperacionErrorArchivoNoSoportado   = MensajeOperacionError + " tipo de archivo no soportado: "
	MensajeOperacionErrorConexion             = MensajeOperacionError + " conexión con la base de datos: "
	MensajeOperacionErrorConsulta             = MensajeOperacionError + " consulta: "
	MensajeOperacionErrorDatosInvalidos       = MensajeOperacionError + " datos inválidos: "
	MensajeOperacionErrorInterno              = MensajeOperacionError + " ocurrió algo inesperado: "
	MensajeOperacionErrorMapeo                = MensajeOperacionError + " mapeando resultados: "
	MensajeOperacionErrorNoAutorizado         = MensajeOperacionError + " acceso no autorizado: "
	MensajeOperacionErrorOperacionNoPermitida = MensajeOperacionError + " operación no permitida: "
	MensajeOperacionErrorParametrosFaltantes  = MensajeOperacionError + " parámetros faltantes: "
	MensajeOperacionErrorParametrosInvalidos  = MensajeOperacionError + " parámetros inválidos: "
	MensajeOperacionErrorPasswordNoEncriptado = MensajeOperacionError + " password no encriptado: "
	MensajeOperacionErrorPasswordNoGenerado   = MensajeOperacionError + " password no generado: "
	MensajeOperacionErrorRecursoNoActualizado = MensajeOperacionError + " registro no actualizado: "
	MensajeOperacionErrorRecursoNoCreado      = MensajeOperacionError + " registro no creado: "
	MensajeOperacionErrorRecursoNoEliminado   = MensajeOperacionError + " registro no eliminado: "
	MensajeOperacionErrorRecursoNoEncontrado  = MensajeOperacionError + " registro no encontrado: "
	MensajeOperacionErrorRecursoRelacion      = MensajeOperacionError + " registro tiene campos relacionados: "
	MensajeOperacionErrorRecursoYaExiste      = MensajeOperacionError + " el registro ya existe: "
	MensajeOperacionErrorSesionExpirada       = MensajeOperacionError + " sesión expirada: "
	MensajeOperacionErrorTiempoEsperaExcedido = MensajeOperacionError + " tiempo de espera excedido: "
	MensajeOperacionErrorTokenFaltante        = MensajeOperacionError + " token de autenticación faltante: "
	MensajeOperacionErrorTokenInvalido        = MensajeOperacionError + " token inválido o expirado: "
	MensajeOperacionErrorUUIDNoGenerado       = MensajeOperacionError + " uuid no generado: "
	MensajeOperacionErrorParseo               = MensajeOperacionError + " parseando datos: "
	MensajeOperacionErrorElementoNoContenido  = MensajeOperacionError + " el elemento no se encuentra dentro de las opciones: "

	// Mensajes Operación Error - PostgreSQL
	MensajeOperacionErrorCampoRequeridoFaltante      = MensajeOperacionError + " campo requerido faltante: "
	MensajeOperacionErrorColumnaDuplicada            = MensajeOperacionError + " columna duplicada: "
	MensajeOperacionErrorColumnaNoExiste             = MensajeOperacionError + " la columna no existe: "
	MensajeOperacionErrorCredencialesInvalidas       = MensajeOperacionError + " credenciales inválidas: "
	MensajeOperacionErrorDatosDemasiadoLargos        = MensajeOperacionError + " datos demasiado largos: "
	MensajeOperacionErrorDemasiadasConexiones        = MensajeOperacionError + " demasiadas conexiones: "
	MensajeOperacionErrorDivisionPorCero             = MensajeOperacionError + " división por cero: "
	MensajeOperacionErrorDiscoLleno                  = MensajeOperacionError + " disco lleno: "
	MensajeOperacionErrorFalloSerializacion          = MensajeOperacionError + " fallo de serialización: "
	MensajeOperacionErrorFaltaMemoria                = MensajeOperacionError + " falta de memoria: "
	MensajeOperacionErrorFormatoFechaInvalido        = MensajeOperacionError + " formato de fecha/hora inválido: "
	MensajeOperacionErrorInterbloqueoDetectado       = MensajeOperacionError + " interbloqueo detectado: "
	MensajeOperacionErrorObjetoNoExiste              = MensajeOperacionError + " el objeto no existe: "
	MensajeOperacionErrorPermisosInsuficientes       = MensajeOperacionError + " permisos insuficientes: "
	MensajeOperacionErrorRepresentacionTextoInvalida = MensajeOperacionError + " representación de texto inválida: "
	MensajeOperacionErrorSintaxisSQL                 = MensajeOperacionError + " error de sintaxis SQL: "
	MensajeOperacionErrorTablaNoExiste               = MensajeOperacionError + " la tabla no existe: "
	MensajeOperacionErrorValorNumericoFueraDeRango   = MensajeOperacionError + " valor numérico fuera de rango: "

	// Mensajes Requeridos
	MensajeRequeridoCampo = " campo " + MensajeRequerido + ": "

	// Mensajes Requeridos Ingresar
	MensajeRequeridoApellidoMaterno = Ingresar + " apellido materno"
	MensajeRequeridoApellidoPaterno = Ingresar + " apellido paterno"
	MensajeRequeridoContrasena      = Ingresar + " contraseña"
	MensajeRequeridoCorreo          = Ingresar + " correo electrónico"
	MensajeRequeridoFechaNacimiento = Ingresar + " fecha de nacimiento"
	MensajeRequeridoNombre          = Ingresar + " nombre"
	MensajeRequeridoTelefono        = Ingresar + " número de teléfono"

	// Mensajes Requeridos Seleccionar
	MensajeRequeridoArchivo = Seleccionar + " archivo"
)
