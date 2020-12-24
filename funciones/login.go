package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/doun/terminal/color"
)

var (
	sesionActiva, sesionRoot bool   = false, false
	idSesion, idGrupo        string = "", ""
)

//EjecutarLogin inicia sesión
func EjecutarLogin(name string, password string, id string) {
	if password == "123" {
		password = "123______"
	}
	if !sesionActiva {
		if name != "" && password != "" && id != "" {

			if IDYaRegistrado(id) {

				if Verificacion := VerificarLogin(name, password, id); Verificacion {
					color.Printf("@{w}Usuario @{w}%v @{w} del grupo @{w}%v ingreso correctamente,.\n", name, idGrupo)
				} else {
					color.Println("@{r}No se pudo loggear, datos incorrectos o usuario no existe.")
				}

			} else {
				color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
			}
		} else {
			color.Println("@{r}	Faltan parámetros obligatorios para la función Login.")
		}
	} else {
		color.Println("@{r}	Hay una sesión activa actualmente.")
	}

}

//VerificarLogin devuelve true o false depende si se pudo loggear
func VerificarLogin(name string, password string, id string) bool {

	NameAux, PathAux := GetDatosPart(id)

	if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

		//LEER Y RECORRER EL MBR
		fileMBR, err2 := os.Open(PathAux)
		if err2 != nil { //validar que no sea nulo.
			panic(err2)
		}
		Disco1 := estructuras.MBR{}
		DiskSize := int(unsafe.Sizeof(Disco1))
		DiskData := leerBytes(fileMBR, DiskSize)
		buffer := bytes.NewBuffer(DiskData)
		err := binary.Read(buffer, binary.BigEndian, &Disco1)
		if err != nil {
			fileMBR.Close()
			fmt.Println(err)
			return false
		}

		//LEER EL SUPERBLOQUE
		InicioParticion := Disco1.Mpartitions[Indice].Pstart
		fileMBR.Seek(int64(InicioParticion+1), 0)
		SB1 := estructuras.Superblock{}
		SBsize := int(unsafe.Sizeof(SB1))
		SBData := leerBytes(fileMBR, SBsize)
		buffer2 := bytes.NewBuffer(SBData)
		err = binary.Read(buffer2, binary.BigEndian, &SB1)
		if err != nil {
			fileMBR.Close()
			fmt.Println(err)
			return false
		}

		if SB1.MontajesCount > 0 {

			//LEEMOS EL PRIMER INODO QUE ES EL INODO DEL ARCHIVO USERS.TXT
			InodoUsers := SB1.InicioInodos
			fileMBR.Seek(int64(InodoUsers+1), 0)
			InodoAux := estructuras.Inodo{}
			InodoSize := int(unsafe.Sizeof(InodoAux))
			InodoData := leerBytes(fileMBR, InodoSize)
			buffer2 := bytes.NewBuffer(InodoData)
			err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
			if err != nil {
				fileMBR.Close()
				fmt.Println(err)
				return false
			}
			//Dividimos el FileSize entre 25 y lo aproximamos al entero a la derecha más cercano
			//de ser necesario, esto para saber cuando bloques ocupa el archivo en total.
			//En caso que un archivo ocupe más de los 4 bloques directos, la cantidad de bytes será mayor a 100
			//y necesitamos crear el slice "Contenido"
			//Por cada bloque usadom leeremos 25 chars
			Resultado := float64(InodoAux.FileSize) / float64(25)
			Resultado = Roundf(Resultado)
			ArrayLen := int(Resultado) * 25
			Contenido := make([]byte, int(ArrayLen))

			//Recorremos los Bloques de Datos del inodo Aux, donde está la información
			//del documento users.txt
			Continuar := true
			x := 0
			for Continuar {
				for i := 0; i < 4; i++ {

					if InodoAux.ApuntadoresBloques[i] > 0 {
						ApuntadorBloque := InodoAux.ApuntadoresBloques[i]
						fileMBR.Seek(int64(ApuntadorBloque+1), 0)
						BloqueAux := estructuras.BloqueDatos{}
						BloqueSize := int(unsafe.Sizeof(BloqueAux))
						BloqueData := leerBytes(fileMBR, BloqueSize)
						buffer := bytes.NewBuffer(BloqueData)
						err = binary.Read(buffer, binary.BigEndian, &BloqueAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return false
						}
						//Esta linea se encarga de ordenar cada bloque (paquete de 25 bytes) en el slice creado
						copy(Contenido[x:x+25], BloqueAux.Data[:])
						x += 25
					}
				}
				//Si el archivo ocupa más de 4 bloques
				//el inodo a punta a otro inodo, asi que lo leemos y seguimos en el Loop
				if InodoAux.ApuntadorIndirecto > 0 {
					fileMBR.Seek(int64(InodoAux.ApuntadorIndirecto+1), 0)
					InodoData := leerBytes(fileMBR, InodoSize)
					buffer2 := bytes.NewBuffer(InodoData)
					err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
					if err != nil {
						fileMBR.Close()
						fmt.Println(err)
						return false
					}
				} else {
					Continuar = false
				}
			}

			ContenidoSize := int(InodoAux.FileSize)
			//EN ESTA PARTE CADENACONTENIDO YA TIENE EL CONTENIDO DE Users.txt
			CadenaContenido := string(Contenido[:ContenidoSize])
			split := strings.Split(CadenaContenido, "\n")
			for _, s := range split {
				registro := strings.Split(s, ",")
				if registro[1] == "U" && registro[0] != "0" {
					if registro[3] == name && registro[4] == password {
						fileMBR.Close()
						idSesion = name
						idGrupo = registro[2]
						sesionActiva = true
						if name == "root" {
							sesionRoot = true
						}
						return true
					}
				}

			}

		} else {
			color.Println("@{r} La partición indicada esta formateada.")
		}

		fileMBR.Close()

	} else if ExisteL, IndiceL := ExisteParticionLogica(PathAux, NameAux); ExisteL {

		fileMBR, err := os.Open(PathAux)
		if err != nil { //validar que no sea nulo.
			panic(err)
		}

		EBRAux := estructuras.EBR{}
		EBRSize := int(unsafe.Sizeof(EBRAux))

		//LEER EL SUPERBLOQUE
		InicioParticion := IndiceL + EBRSize
		fileMBR.Seek(int64(InicioParticion+1), 0)
		SB1 := estructuras.Superblock{}
		SBsize := int(unsafe.Sizeof(SB1))
		SBData := leerBytes(fileMBR, SBsize)
		buffer2 := bytes.NewBuffer(SBData)
		err = binary.Read(buffer2, binary.BigEndian, &SB1)
		if err != nil {
			fileMBR.Close()
			fmt.Println(err)
			return false
		}

		if SB1.MontajesCount > 0 {

			//LEEMOS EL PRIMER INODO QUE ES EL INODO DEL ARCHIVO USERS.TXT
			InodoUsers := SB1.InicioInodos
			fileMBR.Seek(int64(InodoUsers+1), 0)
			InodoAux := estructuras.Inodo{}
			InodoSize := int(unsafe.Sizeof(InodoAux))
			InodoData := leerBytes(fileMBR, InodoSize)
			buffer2 := bytes.NewBuffer(InodoData)
			err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
			if err != nil {
				fileMBR.Close()
				fmt.Println(err)
				return false
			}
			//Dividimos el FileSize entre 25 y lo aproximamos al entero a la derecha más cercano
			//en caso de ser un decimal, para saber cuando bloques ocupa el archivo
			//Por cada bloque usadom leeremos 25 chars
			Resultado := float64(InodoAux.FileSize) / float64(25)
			Resultado = Roundf(Resultado)
			ArrayLen := int(Resultado) * 25
			Contenido := make([]byte, int(ArrayLen))

			//Recorremos los Bloques de Datos del inodo Aux, donde está la información
			//del documento users.txt
			Continuar := true
			x := 0
			for Continuar {
				for i := 0; i < 4; i++ {

					if InodoAux.ApuntadoresBloques[i] > 0 {
						ApuntadorBloque := InodoAux.ApuntadoresBloques[i]
						fileMBR.Seek(int64(ApuntadorBloque+1), 0)
						BloqueAux := estructuras.BloqueDatos{}
						BloqueSize := int(unsafe.Sizeof(BloqueAux))
						BloqueData := leerBytes(fileMBR, BloqueSize)
						buffer := bytes.NewBuffer(BloqueData)
						err = binary.Read(buffer, binary.BigEndian, &BloqueAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return false
						}
						copy(Contenido[x:x+25], BloqueAux.Data[:])
						x += 25
					}
				}

				if InodoAux.ApuntadorIndirecto > 0 {
					fileMBR.Seek(int64(InodoAux.ApuntadorIndirecto+1), 0)
					InodoData := leerBytes(fileMBR, InodoSize)
					buffer2 := bytes.NewBuffer(InodoData)
					err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
					if err != nil {
						fileMBR.Close()
						fmt.Println(err)
						return false
					}
				} else {
					Continuar = false
				}
			}

			ContenidoSize := int(InodoAux.FileSize)
			//EN ESTA PARTE CADENACONTENIDO YA TIENE EL CONTENIDO DE Users.txt
			CadenaContenido := string(Contenido[:ContenidoSize])
			split := strings.Split(CadenaContenido, "\n")
			for _, s := range split {
				registro := strings.Split(s, ",")
				if registro[1] == "U" && registro[0] != "0" {
					if registro[3] == name && registro[4] == password {
						fileMBR.Close()
						idSesion = name
						idGrupo = registro[2]
						sesionActiva = true
						if name == "root" {
							sesionRoot = true
						}
						return true
					}
				}

			}

		} else {
			color.Println("@{r} La partición indicada no esta formateada.")
		}

		fileMBR.Close()

	}
	return false
}

//Roundf proxima un número decimal al entero a la derecha más cercano
func Roundf(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) > 0 {
		return t + math.Copysign(1, x)
	}
	return t
}
