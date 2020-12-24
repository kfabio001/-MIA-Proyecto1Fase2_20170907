package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarRmusr inicia la creación de un nuevo grupo
func EjecutarRmusr(name string, id string) {

	if sesionRoot {

		if name != "" && id != "" {

			if len(name) <= 10 {

				if IDYaRegistrado(id) {

					if name != "root" {

						if UsuarioYaExiste := ExisteUsuario(name, id); UsuarioYaExiste {

							NameAux, PathAux := GetDatosPart(id)

							if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

								fileMBR, err2 := os.OpenFile(PathAux, os.O_RDWR, 0666)
								if err2 != nil {
									fmt.Println(err2)
									fileMBR.Close()
								}

								// Change permissions Linux.
								err2 = os.Chmod(PathAux, 0666)
								if err2 != nil {
									log.Println(err2)
								}

								//Leemos el MBR
								Disco1 := estructuras.MBR{}
								DiskSize := int(unsafe.Sizeof(Disco1))
								DiskData := leerBytes(fileMBR, DiskSize)
								buffer := bytes.NewBuffer(DiskData)
								err := binary.Read(buffer, binary.BigEndian, &Disco1)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
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
									return
								}

								if SB1.MontajesCount > 0 {

									/////////////////////////////////////////////////////////////////////////////////

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
										return
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
													return
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
												return
											}
										} else {
											Continuar = false
										}
									}

									ContenidoSize := int(InodoAux.FileSize)
									//EN ESTA PARTE CADENACONTENIDO YA TIENE EL CONTENIDO DE Users.txt
									CadenaContenido := string(Contenido[:ContenidoSize])
									//Debemos concatenar la informacion del nuevo usuario
									//El formato es: ID,Tipo,Grupo,usuario,contraseña
									// Tambien concatenar un salto de linea al inicio
									ContenidoAux := ""
									split := strings.Split(CadenaContenido, "\n")
									for _, s := range split {
										registro := strings.Split(s, ",")
										if registro[1] == "U" && registro[0] != "0" {
											if registro[3] == name {

												if ContenidoAux == "" {

													ContenidoAux += "0" + "," + registro[1] + "," + registro[2] + "," + registro[3] + "," + registro[4]

												} else {
													ContenidoAux += "\n"
													ContenidoAux += "0" + "," + registro[1] + "," + registro[2] + "," + registro[3] + "," + registro[4]
												}

											} else {

												if ContenidoAux == "" {
													ContenidoAux += s
												} else {
													ContenidoAux += "\n"
													ContenidoAux += s
												}

											}

										} else {

											if ContenidoAux == "" {
												ContenidoAux += s
											} else {
												ContenidoAux += "\n"
												ContenidoAux += s
											}

										}

									}

									CadenaContenido = ContenidoAux

									//Antes de guardar el nuevo contenido debemos setear a cero los apuntadores y las posiciones en el bitmap
									//LEEMOS EL PRIMER INODO QUE ES EL INODO DEL ARCHIVO USERS.TXT
									InodoUsers = SB1.InicioInodos
									fileMBR.Seek(int64(InodoUsers+1), 0)
									InodoAux = estructuras.Inodo{}
									InodoSize = int(unsafe.Sizeof(InodoAux))
									InodoData = leerBytes(fileMBR, InodoSize)
									buffer2 = bytes.NewBuffer(InodoData)
									err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}

									Continuar = true
									x = 0
									for Continuar {
										for i := 0; i < 4; i++ {

											if InodoAux.ApuntadoresBloques[i] > 0 {
												PosicionBitmap := (InodoAux.ApuntadoresBloques[i] - SB1.InicioBloques) / SB1.SizeBloque
												fileMBR.Seek(int64(SB1.InicioBitmapBloques+PosicionBitmap+1), 0)
												data := []byte{0x00}
												fileMBR.Write(data)
												SB1.FreeBloques = SB1.FreeBloques + int32(1)
											}
										}

										if InodoAux.ApuntadorIndirecto > 0 {

											Posicioninodo := (InodoAux.ApuntadorIndirecto - SB1.InicioInodos) / SB1.SizeInodo
											fileMBR.Seek(int64(SB1.InicioBitmapInodos+Posicioninodo+1), 0)
											data := []byte{0x00}
											fileMBR.Write(data)
											SB1.FreeInodos = SB1.FreeInodos + int32(1)

											fileMBR.Seek(int64(InodoAux.ApuntadorIndirecto+1), 0)
											InodoData := leerBytes(fileMBR, InodoSize)
											buffer2 := bytes.NewBuffer(InodoData)
											err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
											if err != nil {
												fileMBR.Close()
												fmt.Println(err)
												return
											}
										} else {
											Continuar = false
										}
									}

									InodoUsers = SB1.InicioInodos
									fileMBR.Seek(int64(InodoUsers+1), 0)
									InodoAux = estructuras.Inodo{}
									InodoSize = int(unsafe.Sizeof(InodoAux))
									InodoData = leerBytes(fileMBR, InodoSize)
									buffer2 = bytes.NewBuffer(InodoData)
									err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}

									//Dividimos la cantidad de caracteres que tiene contenido dentro de 25
									//si el resultado es un numero decimal lo aproximamos al entero más cercano a la derecha
									//esto para saber cuando bloques de datos vamos a usar
									//Ejemplo: Si tenemos 52 caracteres, al dividirlo dentro de 25 obtenemos 2.08
									//ese 0.08 nos obliga a tomar un 3er bloque, entonces la funcion Roundf, aproximaria 2.08 a 3
									FileSize := len(CadenaContenido)
									resultado := float64(len(CadenaContenido)) / float64(25)
									resultado = Roundf(resultado)
									CantidadBloques := int32(resultado)
									//esta variable nos ayudara al corrimiento de los caracteres
									x = 0
									indx := 0
									///////////ESCRIBIMOS LA INFO EN EL INODO

									for i := 0; i < int(resultado); i++ {

										//Creamos un bloque datos, escribimos en su bitmap, creamos el struct, el asignamos los datos desde x hasta x+25
										//si el index llega más de 3, significa que necesitaremos otro inodo

										//Buscamos la posicion en el bitmap para el nuevo BloqueDatos
										PosicionEnBitmapBloque := GetBitmap(fileMBR, int(SB1.InicioBitmapBloques), int(SB1.TotalBloques))
										//Calculamos la posicion del byte en el bitmap BloqueDatos
										BitmapPos := int(SB1.InicioBitmapBloques) + PosicionEnBitmapBloque
										//Escribimos un 1 en esa posición del bitmap
										fileMBR.Seek(int64(BitmapPos+1), 0)
										data := []byte{0x01}
										fileMBR.Write(data)
										//Calculamos la posición del byte del nuevo Bloque Datos
										BloquePos := int(SB1.InicioBloques) + (int(SB1.SizeBloque) * (PosicionEnBitmapBloque))
										//Creamos el nuevo Bloque
										newBloque := estructuras.BloqueDatos{}
										//Le pasamos el contenido desde x hasta x+25

										if i != int(resultado-1) {
											copy(newBloque.Data[:], CadenaContenido[x:x+25])
										} else {
											copy(newBloque.Data[:], CadenaContenido[x:])
										}

										//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
										fileMBR.Seek(int64(BloquePos+1), 0)
										bloquep := &newBloque
										var binario6 bytes.Buffer
										binary.Write(&binario6, binary.BigEndian, bloquep)
										escribirBytes(fileMBR, binario6.Bytes())

										//Actualizamos el SB
										SB1.FreeBloques = SB1.FreeBloques - 1

										//Asignamos al inodo el apuntador al bloque que creamos
										InodoAux.ApuntadoresBloques[indx] = int32(BloquePos)

										x += 25
										indx++
										if indx > 3 && (i+1) < int(resultado) {
											//Si el index es mayor a 3 y estamos seguros que daremos otra iteración más
											//eso significa que necesitamos calcula la posición para un nuevo inodo
											//esa posicion la seteamos en el apuntador indirecto de nuestro newInodo
											//escribimos nuestro newinodo en el archivo, seguido de esto, creamos otro inodo
											//y newInodo apuntaria a un nuevoInodo, finalmente reseteamos indx a 0
											//para que pueda comenzar a apuntar desde la primera posición en el arreglo de apuntadores
											//de bloques del nuevo Inodo

											//Buscamos la posición en el bitmap para el nuevo Inodo
											PosicionEnBitmapInodo := GetBitmap(fileMBR, int(SB1.InicioBitmapInodos), int(SB1.TotalInodos))
											//Calculamos la posicion del byte en el bitmap Inodos
											BitmapPos := int(SB1.InicioBitmapInodos) + PosicionEnBitmapInodo
											//Escribimos un 1 en esa posición del bitmap
											fileMBR.Seek(int64(BitmapPos+1), 0)
											data := []byte{0x01}
											fileMBR.Write(data)
											//Calculamos la posición del byte del nuevo Inodo
											Inodo2Pos := int(SB1.InicioInodos) + (int(SB1.SizeInodo) * (PosicionEnBitmapInodo))

											//Asignamos la posición del nuevo inodo a nuestro newInodo original

											InodoAux.ApuntadorIndirecto = int32(Inodo2Pos)
											InodoAux.FileSize = int32(FileSize)
											//newInodo.NumeroBloques = int32(resultado)
											InodoAux.NumeroBloques = 4
											CantidadBloques = CantidadBloques - 4

											//Ahora toca escribir el struct Inodo en su posición correspondiente
											InodoUsers = SB1.InicioInodos
											fileMBR.Seek(int64(InodoUsers+1), 0)
											inodop := &InodoAux
											var binario bytes.Buffer
											binary.Write(&binario, binary.BigEndian, inodop)
											escribirBytes(fileMBR, binario.Bytes())

											//Ahora creamos el nuevo inodo
											InodoAux = estructuras.Inodo{}
											InodoUsers = int32(Inodo2Pos)

											//Seteando nombre del propietario, en este caso pertenece al id del usuario en curso
											var ArrayProper [20]byte
											nombrePropietario := idSesion
											copy(ArrayProper[:], nombrePropietario)
											copy(InodoAux.Proper[:], ArrayProper[:])
											//Seteando nombre del grupo, en este caso pertenece al id del grupo en curso
											var ArrayGrupo [20]byte
											nombreGrupo := idGrupo
											copy(ArrayGrupo[:], nombreGrupo)
											copy(InodoAux.Grupo[:], ArrayGrupo[:])
											InodoAux.NumeroInodo = int32(int32(SB1.TotalInodos)-int32(SB1.FreeInodos)) + 1
											InodoAux.PermisoU = 7
											InodoAux.PermisoG = 7
											InodoAux.PermisoO = 7

											//Seteamos SB
											SB1.FreeInodos = SB1.FreeInodos - 1

											indx = 0
										}
									}

									InodoAux.FileSize = int32(FileSize)
									InodoAux.NumeroBloques = CantidadBloques

									//Ahora toca escribir el ultimo strcut Inodo creado en su posición correspondiente
									fileMBR.Seek(int64(InodoUsers+1), 0)
									inodop := &InodoAux
									var binario bytes.Buffer
									binary.Write(&binario, binary.BigEndian, inodop)
									escribirBytes(fileMBR, binario.Bytes())

									//Crear bitacora MKUSR
									//Creamos la bitacora para la creación de la carpeta
									BitacoraAux := estructuras.Bitacora{}
									//Seteamos el path, en este caso la primera carpeta tiene "/" como path
									var PathChars [300]byte
									PathAux := "users.txt"
									copy(PathChars[:], PathAux)
									copy(BitacoraAux.Path[:], PathChars[:])
									//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
									var OperacionChars [16]byte
									OperacionAux := "Rmusr"
									copy(OperacionChars[:], OperacionAux)
									copy(BitacoraAux.Operacion[:], OperacionChars[:])
									//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
									BitacoraAux.Tipo = 0
									//Setemos el contenido
									ContenidoRmusr := name
									var ContenidoChars [300]byte
									copy(ContenidoChars[:], ContenidoRmusr)
									copy(BitacoraAux.Contenido[:], ContenidoChars[:])
									BitacoraAux.Size = 0
									//Seteamo la fecha de creación de la bitácora
									t := time.Now()
									var charsTime [20]byte
									cadena := t.Format("2006-01-02 15:04:05")
									copy(charsTime[:], cadena)
									copy(BitacoraAux.Fecha[:], charsTime[:])
									//Calculamos la posicion en la particion donde debemos escribir la bitacora
									NumeroBitacoras := int(SB1.TotalBitacoras - SB1.FreeBitacoras)
									//en este caso al ser la primera bitacora ira al inicio del bloque de bitacoras
									BitacoraPos := int(SB1.InicioBitacora) + (NumeroBitacoras * int(SB1.SizeBitacora))
									//Ahora toca escribir el struct Bitacora en su posición correspondiente
									fileMBR.Seek(int64(BitacoraPos+1), 0)
									bitacorap := &BitacoraAux
									var binario8 bytes.Buffer
									binary.Write(&binario8, binary.BigEndian, bitacorap)
									escribirBytes(fileMBR, binario8.Bytes())

									//Setear nuevas propiedades del superblock
									SB1.FreeBitacoras = SB1.FreeBitacoras - int32(1)
									SB1.FirstFreeInodo = SB1.InicioInodos + (int32(GetBitmap(fileMBR, int(SB1.InicioBitmapInodos), int(SB1.TotalInodos))))
									SB1.FirstFreeBloque = SB1.InicioBloques + (int32(GetBitmap(fileMBR, int(SB1.InicioBitmapBloques), int(SB1.TotalBloques))))

									fileMBR.Seek(int64(InicioParticion+1), 0)
									//Reescribiendo el Superbloque
									sb1 := &SB1
									var binario1 bytes.Buffer
									binary.Write(&binario1, binary.BigEndian, sb1)
									escribirBytes(fileMBR, binario1.Bytes())
									//Reescribir el Backup del Superbloque
									fileMBR.Seek(int64(SB1.InicioBitacora+(SB1.SizeBitacora*SB1.TotalBitacoras)+1), 0)
									sb2 := &SB1
									var binario2 bytes.Buffer
									binary.Write(&binario2, binary.BigEndian, sb2)
									escribirBytes(fileMBR, binario2.Bytes())
									color.Printf("@{w}El usuario @{w}%v @{w}fue eliminado \n", name)

									//EN ESTE MOMENTO EL USUARIO HA SIDO ELIMINADO CORRECTAMENTE, SIN EMBARGO,
									//DEBEMOS ACTUALIZAR TODOS AQUELLAS CARPETAS Y ARCHIVOS CUYO PROPIETARIO ERA EL USUARIO ELIMINADO
									//Y PONERLO EN PROPIEDAD DE ROOT

									//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
									ApuntadorAVD := int(SB1.InicioAVDS)
									//CREAMOS UN STRUCT TEMPORAL
									AVDroot := estructuras.AVD{}
									SizeAVD := int(unsafe.Sizeof(AVDroot))
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									RootData := leerBytes(fileMBR, int(SizeAVD))
									buffer3 := bytes.NewBuffer(RootData)
									err = binary.Read(buffer3, binary.BigEndian, &AVDroot)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}

									CambiarPropietarioAVDRecursivo(fileMBR, ApuntadorAVD, &AVDroot, name, "root", "root")
									fileMBR.Sync()

									/////////////////////////////////////////////////////////////////////////////////

								} else {
									color.Println("@{r} La partición indicada no ha sido formateada.")
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
									return
								}

								if SB1.MontajesCount > 0 {

									/////////////////////////////////////////////////////////////////////////////////

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
										return
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
													return
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
												return
											}
										} else {
											Continuar = false
										}
									}

									ContenidoSize := int(InodoAux.FileSize)
									//EN ESTA PARTE CADENACONTENIDO YA TIENE EL CONTENIDO DE Users.txt
									CadenaContenido := string(Contenido[:ContenidoSize])
									//Debemos concatenar la informacion del nuevo usuario
									//El formato es: ID,Tipo,Grupo,usuario,contraseña
									// Tambien concatenar un salto de linea al inicio
									ContenidoAux := ""
									split := strings.Split(CadenaContenido, "\n")
									for _, s := range split {
										registro := strings.Split(s, ",")
										if registro[1] == "U" && registro[0] != "0" {
											if registro[3] == name {

												if ContenidoAux == "" {

													ContenidoAux += "0" + "," + registro[1] + "," + registro[2] + "," + registro[3] + "," + registro[4]

												} else {
													ContenidoAux += "\n"
													ContenidoAux += "0" + "," + registro[1] + "," + registro[2] + "," + registro[3] + "," + registro[4]
												}

											} else {

												if ContenidoAux == "" {
													ContenidoAux += s
												} else {
													ContenidoAux += "\n"
													ContenidoAux += s
												}

											}

										} else {

											if ContenidoAux == "" {
												ContenidoAux += s
											} else {
												ContenidoAux += "\n"
												ContenidoAux += s
											}

										}

									}

									CadenaContenido = ContenidoAux

									//Antes de guardar el nuevo contenido debemos setear a cero los apuntadores y las posiciones en el bitmap
									//LEEMOS EL PRIMER INODO QUE ES EL INODO DEL ARCHIVO USERS.TXT
									InodoUsers = SB1.InicioInodos
									fileMBR.Seek(int64(InodoUsers+1), 0)
									InodoAux = estructuras.Inodo{}
									InodoSize = int(unsafe.Sizeof(InodoAux))
									InodoData = leerBytes(fileMBR, InodoSize)
									buffer2 = bytes.NewBuffer(InodoData)
									err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}

									Continuar = true
									x = 0
									for Continuar {
										for i := 0; i < 4; i++ {

											if InodoAux.ApuntadoresBloques[i] > 0 {
												PosicionBitmap := (InodoAux.ApuntadoresBloques[i] - SB1.InicioBloques) / SB1.SizeBloque
												fileMBR.Seek(int64(SB1.InicioBitmapBloques+PosicionBitmap+1), 0)
												data := []byte{0x00}
												fileMBR.Write(data)
												SB1.FreeBloques = SB1.FreeBloques + int32(1)
											}
										}

										if InodoAux.ApuntadorIndirecto > 0 {

											Posicioninodo := (InodoAux.ApuntadorIndirecto - SB1.InicioInodos) / SB1.SizeInodo
											fileMBR.Seek(int64(SB1.InicioBitmapInodos+Posicioninodo+1), 0)
											data := []byte{0x00}
											fileMBR.Write(data)
											SB1.FreeInodos = SB1.FreeInodos + int32(1)

											fileMBR.Seek(int64(InodoAux.ApuntadorIndirecto+1), 0)
											InodoData := leerBytes(fileMBR, InodoSize)
											buffer2 := bytes.NewBuffer(InodoData)
											err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
											if err != nil {
												fileMBR.Close()
												fmt.Println(err)
												return
											}
										} else {
											Continuar = false
										}
									}

									InodoUsers = SB1.InicioInodos
									fileMBR.Seek(int64(InodoUsers+1), 0)
									InodoAux = estructuras.Inodo{}
									InodoSize = int(unsafe.Sizeof(InodoAux))
									InodoData = leerBytes(fileMBR, InodoSize)
									buffer2 = bytes.NewBuffer(InodoData)
									err = binary.Read(buffer2, binary.BigEndian, &InodoAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}

									//Dividimos la cantidad de caracteres que tiene contenido dentro de 25
									//si el resultado es un numero decimal lo aproximamos al entero más cercano a la derecha
									//esto para saber cuando bloques de datos vamos a usar
									//Ejemplo: Si tenemos 52 caracteres, al dividirlo dentro de 25 obtenemos 2.08
									//ese 0.08 nos obliga a tomar un 3er bloque, entonces la funcion Roundf, aproximaria 2.08 a 3
									FileSize := len(CadenaContenido)
									resultado := float64(len(CadenaContenido)) / float64(25)
									resultado = Roundf(resultado)
									CantidadBloques := int32(resultado)
									//esta variable nos ayudara al corrimiento de los caracteres
									x = 0
									indx := 0
									///////////ESCRIBIMOS LA INFO EN EL INODO

									for i := 0; i < int(resultado); i++ {

										//Creamos un bloque datos, escribimos en su bitmap, creamos el struct, el asignamos los datos desde x hasta x+25
										//si el index llega más de 3, significa que necesitaremos otro inodo

										//Buscamos la posicion en el bitmap para el nuevo BloqueDatos
										PosicionEnBitmapBloque := GetBitmap(fileMBR, int(SB1.InicioBitmapBloques), int(SB1.TotalBloques))
										//Calculamos la posicion del byte en el bitmap BloqueDatos
										BitmapPos := int(SB1.InicioBitmapBloques) + PosicionEnBitmapBloque
										//Escribimos un 1 en esa posición del bitmap
										fileMBR.Seek(int64(BitmapPos+1), 0)
										data := []byte{0x01}
										fileMBR.Write(data)
										//Calculamos la posición del byte del nuevo Bloque Datos
										BloquePos := int(SB1.InicioBloques) + (int(SB1.SizeBloque) * (PosicionEnBitmapBloque))
										//Creamos el nuevo Bloque
										newBloque := estructuras.BloqueDatos{}
										//Le pasamos el contenido desde x hasta x+25

										if i != int(resultado-1) {
											copy(newBloque.Data[:], CadenaContenido[x:x+25])
										} else {
											copy(newBloque.Data[:], CadenaContenido[x:])
										}

										//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
										fileMBR.Seek(int64(BloquePos+1), 0)
										bloquep := &newBloque
										var binario6 bytes.Buffer
										binary.Write(&binario6, binary.BigEndian, bloquep)
										escribirBytes(fileMBR, binario6.Bytes())

										//Actualizamos el SB
										SB1.FreeBloques = SB1.FreeBloques - 1

										//Asignamos al inodo el apuntador al bloque que creamos
										InodoAux.ApuntadoresBloques[indx] = int32(BloquePos)

										x += 25
										indx++
										if indx > 3 && (i+1) < int(resultado) {
											//Si el index es mayor a 3 y estamos seguros que daremos otra iteración más
											//eso significa que necesitamos calcula la posición para un nuevo inodo
											//esa posicion la seteamos en el apuntador indirecto de nuestro newInodo
											//escribimos nuestro newinodo en el archivo, seguido de esto, creamos otro inodo
											//y newInodo apuntaria a un nuevoInodo, finalmente reseteamos indx a 0
											//para que pueda comenzar a apuntar desde la primera posición en el arreglo de apuntadores
											//de bloques del nuevo Inodo

											//Buscamos la posición en el bitmap para el nuevo Inodo
											PosicionEnBitmapInodo := GetBitmap(fileMBR, int(SB1.InicioBitmapInodos), int(SB1.TotalInodos))
											//Calculamos la posicion del byte en el bitmap Inodos
											BitmapPos := int(SB1.InicioBitmapInodos) + PosicionEnBitmapInodo
											//Escribimos un 1 en esa posición del bitmap
											fileMBR.Seek(int64(BitmapPos+1), 0)
											data := []byte{0x01}
											fileMBR.Write(data)
											//Calculamos la posición del byte del nuevo Inodo
											Inodo2Pos := int(SB1.InicioInodos) + (int(SB1.SizeInodo) * (PosicionEnBitmapInodo))

											//Asignamos la posición del nuevo inodo a nuestro newInodo original

											InodoAux.ApuntadorIndirecto = int32(Inodo2Pos)
											InodoAux.FileSize = int32(FileSize)
											//newInodo.NumeroBloques = int32(resultado)
											InodoAux.NumeroBloques = 4
											CantidadBloques = CantidadBloques - 4

											//Ahora toca escribir el struct Inodo en su posición correspondiente
											InodoUsers = SB1.InicioInodos
											fileMBR.Seek(int64(InodoUsers+1), 0)
											inodop := &InodoAux
											var binario bytes.Buffer
											binary.Write(&binario, binary.BigEndian, inodop)
											escribirBytes(fileMBR, binario.Bytes())

											//Ahora creamos el nuevo inodo
											InodoAux = estructuras.Inodo{}
											InodoUsers = int32(Inodo2Pos)

											//Seteando nombre del propietario, en este caso pertenece al id del usuario en curso
											var ArrayProper [20]byte
											nombrePropietario := idSesion
											copy(ArrayProper[:], nombrePropietario)
											copy(InodoAux.Proper[:], ArrayProper[:])
											//Seteando nombre del grupo, en este caso pertenece al id del grupo en curso
											var ArrayGrupo [20]byte
											nombreGrupo := idGrupo
											copy(ArrayGrupo[:], nombreGrupo)
											copy(InodoAux.Grupo[:], ArrayGrupo[:])
											InodoAux.NumeroInodo = int32(int32(SB1.TotalInodos)-int32(SB1.FreeInodos)) + 1
											InodoAux.PermisoU = 7
											InodoAux.PermisoG = 7
											InodoAux.PermisoO = 7

											//Seteamos SB
											SB1.FreeInodos = SB1.FreeInodos - 1

											indx = 0
										}
									}

									InodoAux.FileSize = int32(FileSize)
									InodoAux.NumeroBloques = CantidadBloques

									//Ahora toca escribir el ultimo strcut Inodo creado en su posición correspondiente
									fileMBR.Seek(int64(InodoUsers+1), 0)
									inodop := &InodoAux
									var binario bytes.Buffer
									binary.Write(&binario, binary.BigEndian, inodop)
									escribirBytes(fileMBR, binario.Bytes())

									//Crear bitacora MKUSR
									//Creamos la bitacora para la creación de la carpeta
									BitacoraAux := estructuras.Bitacora{}
									//Seteamos el path, en este caso la primera carpeta tiene "/" como path
									var PathChars [300]byte
									PathAux := "users.txt"
									copy(PathChars[:], PathAux)
									copy(BitacoraAux.Path[:], PathChars[:])
									//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
									var OperacionChars [16]byte
									OperacionAux := "Rmusr"
									copy(OperacionChars[:], OperacionAux)
									copy(BitacoraAux.Operacion[:], OperacionChars[:])
									//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
									BitacoraAux.Tipo = 0
									//Setemos el contenido
									ContenidoMkuser := name
									var ContenidoChars [300]byte
									copy(ContenidoChars[:], ContenidoMkuser)
									copy(BitacoraAux.Contenido[:], ContenidoChars[:])
									BitacoraAux.Size = 0
									//Seteamo la fecha de creación de la bitácora
									t := time.Now()
									var charsTime [20]byte
									cadena := t.Format("2006-01-02 15:04:05")
									copy(charsTime[:], cadena)
									copy(BitacoraAux.Fecha[:], charsTime[:])
									//Calculamos la posicion en la particion donde debemos escribir la bitacora
									NumeroBitacoras := int(SB1.TotalBitacoras - SB1.FreeBitacoras)
									//en este caso al ser la primera bitacora ira al inicio del bloque de bitacoras
									BitacoraPos := int(SB1.InicioBitacora) + (NumeroBitacoras * int(SB1.SizeBitacora))
									//Ahora toca escribir el struct Bitacora en su posición correspondiente
									fileMBR.Seek(int64(BitacoraPos+1), 0)
									bitacorap := &BitacoraAux
									var binario8 bytes.Buffer
									binary.Write(&binario8, binary.BigEndian, bitacorap)
									escribirBytes(fileMBR, binario8.Bytes())

									//Setear nuevas propiedades del superblock
									SB1.FreeBitacoras = SB1.FreeBitacoras - int32(1)
									SB1.FirstFreeInodo = SB1.InicioInodos + (int32(GetBitmap(fileMBR, int(SB1.InicioBitmapInodos), int(SB1.TotalInodos))))
									SB1.FirstFreeBloque = SB1.InicioBloques + (int32(GetBitmap(fileMBR, int(SB1.InicioBitmapBloques), int(SB1.TotalBloques))))

									fileMBR.Seek(int64(InicioParticion+1), 0)
									//Reescribiendo el Superbloque
									sb1 := &SB1
									var binario1 bytes.Buffer
									binary.Write(&binario1, binary.BigEndian, sb1)
									escribirBytes(fileMBR, binario1.Bytes())
									//Reescribir el Backup del Superbloque
									fileMBR.Seek(int64(SB1.InicioBitacora+(SB1.SizeBitacora*SB1.TotalBitacoras)+1), 0)
									sb2 := &SB1
									var binario2 bytes.Buffer
									binary.Write(&binario2, binary.BigEndian, sb2)
									escribirBytes(fileMBR, binario2.Bytes())
									color.Printf("@{w}El usuario @{w}%v @{w}fue eliminado\n", name)

									//EN ESTE MOMENTO EL USUARIO HA SIDO ELIMINADO CORRECTAMENTE, SIN EMBARGO,
									//DEBEMOS ACTUALIZAR TODOS AQUELLAS CARPETAS Y ARCHIVOS CUYO PROPIETARIO ERA EL USUARIO ELIMINADO
									//Y PONERLO EN PROPIEDAD DE ROOT

									//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
									ApuntadorAVD := int(SB1.InicioAVDS)
									//CREAMOS UN STRUCT TEMPORAL
									AVDroot := estructuras.AVD{}
									SizeAVD := int(unsafe.Sizeof(AVDroot))
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									RootData := leerBytes(fileMBR, int(SizeAVD))
									buffer3 := bytes.NewBuffer(RootData)
									err = binary.Read(buffer3, binary.BigEndian, &AVDroot)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}

									CambiarPropietarioAVDRecursivo(fileMBR, ApuntadorAVD, &AVDroot, name, "root", "root")
									fileMBR.Sync()

									/////////////////////////////////////////////////////////////////////////////////

								} else {
									color.Println("@{r} La partición indicada no ha sido formateada.")
								}

								fileMBR.Close()

							}

						} else {
							color.Printf("@{r}No existe ningun usuario registrado con el nombre @{w}%v\n", name)
						}

					} else {
						color.Printf("@{r}El sistema no permite eliminar el usuario @{w}%v\n", name)
					}

				} else {
					color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
				}

			} else {
				color.Println("@{r}	El nombre no puede tener más de 10 caracteres.")
			}

		} else {
			color.Println("@{r}	Faltan parámetros obligatorios para la función Rmusr.")
		}

	} else {
		color.Println("@{r}Se necesita de una sesión root activa para ejecutar la función RMUSR.")
	}

}

//CambiarPropietarioAVDRecursivo recorre un AVD
func CambiarPropietarioAVDRecursivo(file *os.File, ByteAVD int, AVDAux *estructuras.AVD, Actual string, Nuevo string, NuevoGrupo string) {

	var ProperName [20]byte
	copy(ProperName[:], Actual)

	if string(AVDAux.Proper[:]) == string(ProperName[:]) {

		var NewProper [20]byte
		copy(NewProper[:], Nuevo)

		var NewGroup [20]byte
		copy(NewGroup[:], NuevoGrupo)

		copy(AVDAux.Proper[:], NewProper[:])
		copy(AVDAux.Grupo[:], NewGroup[:])

		//Reescribir el AVD
		file.Seek(int64(ByteAVD+1), 0)
		avdp := AVDAux
		var binarioF bytes.Buffer
		binary.Write(&binarioF, binary.BigEndian, avdp)
		escribirBytes(file, binarioF.Bytes())
		file.Sync()
	}

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

			ApuntadorAVD := int(AVDAux.ApuntadorSubs[i])

			//Con el valor del apuntador leemos un struct AVD
			AVDHijo := estructuras.AVD{}
			file.Seek(int64(int32(AVDAux.ApuntadorSubs[i])+int32(1)), 0)
			SizeAVD := int(unsafe.Sizeof(AVDHijo))
			HijoData := leerBytes(file, int(SizeAVD))
			buffer := bytes.NewBuffer(HijoData)
			err := binary.Read(buffer, binary.BigEndian, &AVDHijo)
			if err != nil {
				log.Fatal(err)
				fmt.Println(err)
				return

			}

			CambiarPropietarioAVDRecursivo(file, ApuntadorAVD, &AVDHijo, Actual, Nuevo, NuevoGrupo)

		}
	}

	DDAPuntador := int(AVDAux.ApuntadorDD)
	//Con el valor del apuntador leemos un struct DD
	DDAux := estructuras.DD{}
	_, err := file.Seek(int64(AVDAux.ApuntadorDD+int32(1)), 0)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
		return

	}
	SizeDD := int(unsafe.Sizeof(DDAux))
	DDData := leerBytes(file, int(SizeDD))
	buffer := bytes.NewBuffer(DDData)
	err = binary.Read(buffer, binary.BigEndian, &DDAux)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
		return

	}

	CambiarPropietarioDDRecursivo(file, DDAPuntador, &DDAux, Actual, Nuevo, NuevoGrupo)

	if AVDAux.ApuntadorAVD > 0 {

		ApuntadorAVD := int(AVDAux.ApuntadorAVD)

		//Con el valor del apuntador leemos un struct AVD
		AVDExt := estructuras.AVD{}
		file.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
		SizeAVD := int(unsafe.Sizeof(AVDExt))
		AVDData := leerBytes(file, int(SizeAVD))
		buffer := bytes.NewBuffer(AVDData)
		err := binary.Read(buffer, binary.BigEndian, &AVDExt)
		if err != nil {
			fmt.Println(err)
			return
		}

		CambiarPropietarioAVDExtRecursivo(file, ApuntadorAVD, &AVDExt, Actual, Nuevo, NuevoGrupo)
	}

}

//CambiarPropietarioAVDExtRecursivo recorre la extensión del AVD
func CambiarPropietarioAVDExtRecursivo(file *os.File, ByteAVD int, AVDAux *estructuras.AVD, Actual string, Nuevo string, NuevoGrupo string) {

	var ProperName [20]byte
	copy(ProperName[:], Actual)

	if string(AVDAux.Proper[:]) == string(ProperName[:]) {

		var NewProper [20]byte
		copy(NewProper[:], Nuevo)

		var NewGroup [20]byte
		copy(NewGroup[:], NuevoGrupo)

		copy(AVDAux.Proper[:], NewProper[:])
		copy(AVDAux.Grupo[:], NewGroup[:])

		//Reescribir el AVD
		file.Seek(int64(ByteAVD+1), 0)
		avdp := AVDAux
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, avdp)
		escribirBytes(file, binario2.Bytes())
		file.Sync()
	}

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

			ApuntadorAVD := int(AVDAux.ApuntadorSubs[i])

			//Con el valor del apuntador leemos un struct AVD
			AVDHijo := estructuras.AVD{}
			file.Seek(int64(AVDAux.ApuntadorSubs[i]+int32(1)), 0)
			SizeAVD := int(unsafe.Sizeof(AVDHijo))
			HijoData := leerBytes(file, int(SizeAVD))
			buffer := bytes.NewBuffer(HijoData)
			err := binary.Read(buffer, binary.BigEndian, &AVDHijo)
			if err != nil {
				log.Fatal(err)
				fmt.Println(err)
				return

			}

			CambiarPropietarioAVDRecursivo(file, ApuntadorAVD, &AVDHijo, Actual, Nuevo, NuevoGrupo)

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

		ApuntadorAVD := int(AVDAux.ApuntadorAVD)

		//Con el valor del apuntador leemos un struct AVD
		AVDExt := estructuras.AVD{}
		file.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
		SizeAVD := int(unsafe.Sizeof(AVDExt))
		AVDData := leerBytes(file, int(SizeAVD))
		buffer := bytes.NewBuffer(AVDData)
		err := binary.Read(buffer, binary.BigEndian, &AVDExt)
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
			return
		}

		CambiarPropietarioAVDExtRecursivo(file, ApuntadorAVD, &AVDExt, Actual, Nuevo, NuevoGrupo)
	}

}

//CambiarPropietarioDDRecursivo recorre el detalle de directorio
func CambiarPropietarioDDRecursivo(file *os.File, ByteDD int, DDaux *estructuras.DD, Actual string, Nuevo string, NuevoGrupo string) {

	for i := 0; i < 5; i++ {

		if DDaux.DDFiles[i].ApuntadorInodo > 0 {

			InodoApuntador := int(DDaux.DDFiles[i].ApuntadorInodo)
			//Con el valor del apuntador leemos un struct Inodo
			InodoAux := estructuras.Inodo{}
			file.Seek(int64(DDaux.DDFiles[i].ApuntadorInodo+int32(1)), 0)
			SizeInodo := int(unsafe.Sizeof(InodoAux))
			InodoData := leerBytes(file, int(SizeInodo))
			buffer := bytes.NewBuffer(InodoData)
			err := binary.Read(buffer, binary.BigEndian, &InodoAux)
			if err != nil {
				fmt.Println(err)
				return

			}

			CambiarPropietarioInodoRecursivo(file, InodoApuntador, &InodoAux, Actual, Nuevo, NuevoGrupo)

		}
	}

	if DDaux.ApuntadorDD > 0 {

		DDAPuntador := int(DDaux.ApuntadorDD)

		//Con el valor del apuntador leemos un struct DD
		DDExt := estructuras.DD{}
		file.Seek(int64(DDaux.ApuntadorDD+int32(1)), 0)
		SizeDD := int(unsafe.Sizeof(DDExt))
		ExtData := leerBytes(file, int(SizeDD))
		buffer := bytes.NewBuffer(ExtData)
		err := binary.Read(buffer, binary.BigEndian, &DDExt)
		if err != nil {
			fmt.Println(err)
			return

		}

		CambiarPropietarioDDRecursivo(file, DDAPuntador, &DDExt, Actual, Nuevo, NuevoGrupo)
	}

}

//CambiarPropietarioInodoRecursivo recorre el inodo
func CambiarPropietarioInodoRecursivo(file *os.File, ByteInodo int, InodoAux *estructuras.Inodo, Actual string, Nuevo string, NuevoGrupo string) {

	var ProperName [20]byte
	copy(ProperName[:], Actual)

	if string(InodoAux.Proper[:]) == string(ProperName[:]) {

		var NewProper [20]byte
		copy(NewProper[:], Nuevo)

		var NewGroup [20]byte
		copy(NewGroup[:], NuevoGrupo)

		copy(InodoAux.Proper[:], NewProper[:])
		copy(InodoAux.Grupo[:], NewGroup[:])

		//Reescribir el AVD
		file.Seek(int64(ByteInodo+1), 0)
		inodop := InodoAux
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, inodop)
		escribirBytes(file, binario2.Bytes())
		file.Sync()
	}

	if InodoAux.ApuntadorIndirecto > 0 {

		InodoApuntador := int(InodoAux.ApuntadorIndirecto)

		//Con el valor del apuntador leemos un struct Inodo
		InodoExt := estructuras.Inodo{}
		file.Seek(int64(InodoAux.ApuntadorIndirecto+int32(1)), 0)
		SizeInodo := int(unsafe.Sizeof(InodoExt))
		ExtData := leerBytes(file, int(SizeInodo))
		buffer := bytes.NewBuffer(ExtData)
		err := binary.Read(buffer, binary.BigEndian, &InodoExt)
		if err != nil {
			fmt.Println(err)
			return

		}

		CambiarPropietarioInodoRecursivo(file, InodoApuntador, &InodoExt, Actual, Nuevo, NuevoGrupo)
	}

}
