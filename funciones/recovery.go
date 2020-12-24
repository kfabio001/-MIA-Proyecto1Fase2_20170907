package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarRecovery function
func EjecutarRecovery(id string) {

	if sesionRoot {

		if id != "" {

			if IDYaRegistrado(id) {

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

						//Antes de reescribir el super bloque y su back up debemos hacer un copia temporal del bloque bitacoras y del Backup del super bloque

						fileMBR.Seek(int64(SB1.InicioBitacora+1), 0)
						CantidadBytes := int(SB1.TotalBitacoras*SB1.SizeBitacora) + SBsize
						//Toda la info de las bitacoras y el backUp se almacena en BackupData
						BackUpData := leerBytes(fileMBR, CantidadBytes)
						//Creamos una copia del SuperBloque (generada en LOSS)
						//Nos posicionamos al inicio del backup del superbloque
						fileMBR.Seek(int64(SB1.InicioBitacora+(SB1.TotalBitacoras*SB1.SizeBitacora)+1), 0)
						SBBackUp := leerBytes(fileMBR, int(SBsize))
						//Calculamos el numero de bitacoras que tenamos originalmente, antes de ejecutar LOSS
						//estos atributos del super bloque no se modificaron asi que sigue guardando la cantidad correcta
						NumeroBitacoras := int(SB1.TotalBitacoras - SB1.FreeBitacoras)

						//Primer bitacora:= InicioBitacora + (sizeBitacora * 0)
						//Segunda bitacora := InicioBitacora + (sizeBitacora * 1)
						//Tercer bitacora := InicioBitcora + (sizeBitacora *2)
						//Iteramos desde la 3era bitacora (si existisera), las primeras 2 bitacoras son carpeta Root y users.txt que siempre existen

						ArregloBitacoras := []*estructuras.Bitacora{}
						//Recorremos todas las bitacoras existentes
						for i := 2; i < NumeroBitacoras; i++ {

							PosicionBitacora := int(SB1.InicioBitacora + (SB1.SizeBitacora * int32(i)))
							fileMBR.Seek(int64(PosicionBitacora+1), 0)
							BitacoraAux := estructuras.Bitacora{}
							BitacoraData := leerBytes(fileMBR, int(SB1.SizeBitacora))
							buffer := bytes.NewBuffer(BitacoraData)
							err := binary.Read(buffer, binary.BigEndian, &BitacoraAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}
							//Almacenamos las bitacoras en un arreglo
							ArregloBitacoras = append(ArregloBitacoras, &BitacoraAux)

						}
						//Cerramos el archivo porque las funciones mkdir y mkfile tambien lo abren
						fileMBR.Close()
						//Si guardamos al menos una bitacora la recorremos en este ciclo for
						for x := 0; x < len(ArregloBitacoras); x++ {

							var OperacionAux [16]byte
							cadena := "Mkdir"
							copy(OperacionAux[:], cadena)
							//Es una bitacora tipo Mkdir
							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux[:]) {

								var PathAux [300]byte
								copy(PathAux[:], ArregloBitacoras[x].Path[:])
								n := bytes.Index(PathAux[:], []byte{0})
								if n == -1 {
									n = len(PathAux)
								}
								CadenaPath := string(PathAux[:n])

								var ProperAux [16]byte
								copy(ProperAux[:], ArregloBitacoras[x].Proper[:])
								n = bytes.Index(ProperAux[:], []byte{0})
								if n == -1 {
									n = len(ProperAux)
								}
								PropietarioAux := string(ProperAux[:n])

								var GrupoAux [16]byte
								copy(GrupoAux[:], ArregloBitacoras[x].Grupo[:])
								n = bytes.Index(GrupoAux[:], []byte{0})
								if n == -1 {
									n = len(GrupoAux)
								}
								NombreGrupoAux := string(GrupoAux[:n])

								SesionReal := idSesion
								GrupoReal := idGrupo

								idSesion = PropietarioAux
								idGrupo = NombreGrupoAux

								EjecutarMkdir(id, CadenaPath, "-P")

								idSesion = SesionReal
								idGrupo = GrupoReal
							}

							var OperacionAux2 [16]byte
							cadena2 := "Mkfile"
							copy(OperacionAux2[:], cadena2)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux2[:]) {

								var PathAux [300]byte
								copy(PathAux[:], ArregloBitacoras[x].Path[:])
								n := bytes.Index(PathAux[:], []byte{0})
								if n == -1 {
									n = len(PathAux)
								}
								CadenaPath := string(PathAux[:n])

								ValorSize := fmt.Sprint(ArregloBitacoras[x].Size)

								var contenidoAux [300]byte
								copy(contenidoAux[:], ArregloBitacoras[x].Contenido[:])
								n = bytes.Index(contenidoAux[:], []byte{0})
								if n == -1 {
									n = len(contenidoAux)
								}
								CadenaContenido := string(contenidoAux[:n])

								var ProperAux [16]byte
								copy(ProperAux[:], ArregloBitacoras[x].Proper[:])
								n = bytes.Index(ProperAux[:], []byte{0})
								if n == -1 {
									n = len(ProperAux)
								}
								PropietarioAux := string(ProperAux[:n])

								var GrupoAux [16]byte
								copy(GrupoAux[:], ArregloBitacoras[x].Grupo[:])
								n = bytes.Index(GrupoAux[:], []byte{0})
								if n == -1 {
									n = len(GrupoAux)
								}
								NombreGrupoAux := string(GrupoAux[:n])

								SesionReal := idSesion
								GrupoReal := idGrupo

								idSesion = PropietarioAux
								idGrupo = NombreGrupoAux

								EjecutarMkfile(id, CadenaPath, ValorSize, CadenaContenido, "-P")

								idSesion = SesionReal
								idGrupo = GrupoReal
							}

							var OperacionAux3 [16]byte
							cadena3 := "Mkusr"
							copy(OperacionAux3[:], cadena3)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux3[:]) {

								n := bytes.Index(ArregloBitacoras[x].Contenido[:], []byte{0})
								if n == -1 {
									n = len(ArregloBitacoras[x].Contenido[:])
								}
								Contenido := string(ArregloBitacoras[x].Contenido[:n])
								Atributos := strings.Split(Contenido, ",")
								NameAux := Atributos[0]
								PassAux := Atributos[1]
								GrupoAux := Atributos[2]

								EjecutarMkusr(NameAux, PassAux, GrupoAux, id)

							}

							var OperacionAux4 [16]byte
							cadena4 := "Mkgrp"
							copy(OperacionAux4[:], cadena4)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux4[:]) {

								n := bytes.Index(ArregloBitacoras[x].Contenido[:], []byte{0})
								if n == -1 {
									n = len(ArregloBitacoras[x].Contenido[:])
								}
								Contenido := string(ArregloBitacoras[x].Contenido[:n])

								EjecutarMkgrp(Contenido, id)

							}

							var OperacionAux5 [16]byte
							cadena5 := "Ren"
							copy(OperacionAux5[:], cadena5)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux5[:]) {

								var PathAux [300]byte
								copy(PathAux[:], ArregloBitacoras[x].Path[:])
								n := bytes.Index(PathAux[:], []byte{0})
								if n == -1 {
									n = len(PathAux)
								}
								CadenaPath := string(PathAux[:n])

								var contenidoAux [300]byte
								copy(contenidoAux[:], ArregloBitacoras[x].Contenido[:])
								n = bytes.Index(contenidoAux[:], []byte{0})
								if n == -1 {
									n = len(contenidoAux)
								}
								CadenaContenido := string(contenidoAux[:n])

								EjecutarRen(id, CadenaPath, CadenaContenido)

							}

							var OperacionAux6 [16]byte
							cadena6 := "Rmusr"
							copy(OperacionAux6[:], cadena6)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux6[:]) {

								n := bytes.Index(ArregloBitacoras[x].Contenido[:], []byte{0})
								if n == -1 {
									n = len(ArregloBitacoras[x].Contenido[:])
								}
								Contenido := string(ArregloBitacoras[x].Contenido[:n])

								EjecutarRmusr(Contenido, id)

							}

							var OperacionAux7 [16]byte
							cadena7 := "Chgrp"
							copy(OperacionAux7[:], cadena7)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux7[:]) {

								n := bytes.Index(ArregloBitacoras[x].Contenido[:], []byte{0})
								if n == -1 {
									n = len(ArregloBitacoras[x].Contenido[:])
								}
								Contenido := string(ArregloBitacoras[x].Contenido[:n])
								Atributos := strings.Split(Contenido, ",")
								NameAux := Atributos[0]
								GrupoAux := Atributos[1]

								EjecutarChgrp(NameAux, GrupoAux, id)

							}

							var OperacionAux8 [16]byte
							cadena8 := "Rmgrp"
							copy(OperacionAux8[:], cadena8)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux8[:]) {

								n := bytes.Index(ArregloBitacoras[x].Contenido[:], []byte{0})
								if n == -1 {
									n = len(ArregloBitacoras[x].Contenido[:])
								}
								Contenido := string(ArregloBitacoras[x].Contenido[:n])

								EjecutarRmgrp(Contenido, id)

							}

							var OperacionAux9 [16]byte
							cadena9 := "Chmod"
							copy(OperacionAux9[:], cadena9)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux9[:]) {

								n := bytes.Index(ArregloBitacoras[x].Contenido[:], []byte{0})
								if n == -1 {
									n = len(ArregloBitacoras[x].Contenido[:])
								}
								Contenido := string(ArregloBitacoras[x].Contenido[:n])
								Atributos := strings.Split(Contenido, ",")
								UGOaux := Atributos[0]
								Raux := Atributos[1]

								var PathAux [300]byte
								copy(PathAux[:], ArregloBitacoras[x].Path[:])
								n = bytes.Index(PathAux[:], []byte{0})
								if n == -1 {
									n = len(PathAux)
								}
								CadenaPath := string(PathAux[:n])

								var ProperAux [16]byte
								copy(ProperAux[:], ArregloBitacoras[x].Proper[:])
								n = bytes.Index(ProperAux[:], []byte{0})
								if n == -1 {
									n = len(ProperAux)
								}
								PropietarioAux := string(ProperAux[:n])

								var GrupoAux [16]byte
								copy(GrupoAux[:], ArregloBitacoras[x].Grupo[:])
								n = bytes.Index(GrupoAux[:], []byte{0})
								if n == -1 {
									n = len(GrupoAux)
								}
								NombreGrupoAux := string(GrupoAux[:n])

								SesionReal := idSesion
								GrupoReal := idGrupo

								idSesion = PropietarioAux
								idGrupo = NombreGrupoAux

								EjecutarChmod(id, CadenaPath, UGOaux, Raux)

								idSesion = SesionReal
								idGrupo = GrupoReal

							}

							var OperacionAux10 [16]byte
							cadena10 := "Edit"
							copy(OperacionAux10[:], cadena10)

							if string(ArregloBitacoras[x].Operacion[:]) == string(OperacionAux10[:]) {

								var PathAux [300]byte
								copy(PathAux[:], ArregloBitacoras[x].Path[:])
								n := bytes.Index(PathAux[:], []byte{0})
								if n == -1 {
									n = len(PathAux)
								}
								CadenaPath := string(PathAux[:n])

								ValorSize := fmt.Sprint(ArregloBitacoras[x].Size)

								var contenidoAux [300]byte
								copy(contenidoAux[:], ArregloBitacoras[x].Contenido[:])
								n = bytes.Index(contenidoAux[:], []byte{0})
								if n == -1 {
									n = len(contenidoAux)
								}
								CadenaContenido := string(contenidoAux[:n])

								var ProperAux [16]byte
								copy(ProperAux[:], ArregloBitacoras[x].Proper[:])
								n = bytes.Index(ProperAux[:], []byte{0})
								if n == -1 {
									n = len(ProperAux)
								}
								PropietarioAux := string(ProperAux[:n])

								var GrupoAux [16]byte
								copy(GrupoAux[:], ArregloBitacoras[x].Grupo[:])
								n = bytes.Index(GrupoAux[:], []byte{0})
								if n == -1 {
									n = len(GrupoAux)
								}
								NombreGrupoAux := string(GrupoAux[:n])

								SesionReal := idSesion
								GrupoReal := idGrupo

								idSesion = PropietarioAux
								idGrupo = NombreGrupoAux

								EjecutarEdit(id, CadenaPath, ValorSize, CadenaContenido)

								idSesion = SesionReal
								idGrupo = GrupoReal
							}

						}
						//Volvemos a abrir el Disco en esta parte del codigo para
						//poder escribir, Superbloque, bitacoras originales y backup original
						fileMBR, err2 = os.OpenFile(PathAux, os.O_RDWR, 0666)
						if err2 != nil {
							fmt.Println(err2)
							fileMBR.Close()
						}

						// Change permissions Linux.
						err2 = os.Chmod(PathAux, 0666)
						if err2 != nil {
							log.Println(err2)
						}

						//Escribiendo el BackUp del super bloque, generado en LOSS
						//al inicio de la particion
						fileMBR.Seek(int64(SB1.PartStart+1), 0)
						fileMBR.Write(SBBackUp)
						//Escribiendo el Bloque bitacoras original y el backup del super bloque original
						fileMBR.Seek(int64(SB1.InicioBitacora+1), 0)
						fileMBR.Write(BackUpData)

						color.Println("@{c} Sistema recuperado con exito.")

					} else {
						color.Println("@{r} La partición indicada no ha sido formateada.")
					}

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

					} else {
						color.Println("@{r} La partición indicada no ha sido formateada.")
					}

					fileMBR.Close()

				}

			} else {
				color.Printf("@{r}No hay ninguna partición montada con el id: @{y}%v\n", id)
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios para la funcion RECOVERY.")
		}

	} else {
		color.Println("@{r}Se necesita de una sesión root activa para ejecutar la función RECOVERY.")
	}

}
