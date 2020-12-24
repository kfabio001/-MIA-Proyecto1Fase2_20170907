package funciones

import "github.com/doun/terminal/color"

//EjecutarUnmount function
func EjecutarUnmount(lista *[]string) {

	for _, id := range *lista {

		if IDYaRegistrado(id) {

			NameAux, PathAux := GetDatosPart(id)
			Desmontar(PathAux, NameAux, id)

		} else {
			color.Printf("@{r}No hay ninguna partición montada con el id: %v\n", id)
		}

	}
}

//Desmontar desmonta una particion
func Desmontar(PathAux string, NameAux string, id string) {

	CambiarStatusU(PathAux, NameAux)

	for i, pm := range PMList {
		if pm.PMid == id {
			PMList = append(PMList[:i], PMList[i+1:]...)
		}
	}

	if len(Discos) > 0 {
		for i := 0; i < len(Discos); i++ {
			if Discos[i].MDpath == PathAux {

				for x, part := range Discos[i].Particiones {
					if part == NameAux {
						Discos[i].Particiones = append(Discos[i].Particiones[:x], Discos[i].Particiones[x+1:]...)
					}
				}

			}
		}
	}
}
