package funciones

import "github.com/doun/terminal/color"

//EjecutarLogout termina la sesión en caso que si haya una sesión activa
func EjecutarLogout() {

	if sesionActiva {
		sesionActiva = false
		sesionRoot = false
		idSesion = ""
		idGrupo = ""
		color.Println("@{w}	Sesión cerrada correctamente.")
	} else {
		color.Println("@{r}	No hay ninguna sesión activa actualmente.")
	}

}
