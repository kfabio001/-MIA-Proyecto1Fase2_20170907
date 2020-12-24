package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unsafe"

	"github.com/doun/terminal/color"
)

var (
	cadenaConsola string = ""
)

//EjecutarFind function
func EjecutarFind(id string, path string, name string) {

	if sesionActiva {

		if path != "" && id != "" && name != "" {

			if strings.HasPrefix(path, "/") {

				if path != "/" { // si no es root quita slash al final
					if last := len(path) - 1; last >= 0 && path[last] == '/' {
						path = path[:last]
					}
				}

				if len(name) <= 20 {

					if name == "*" {
						FindAll(id, path, name)
					} else if name == "?.*" { //nombre del archivo un caracter y cualquier extension
						FindCaso1(id, path, name)
					} else {
						FindCaso2(id, path, name)
					}

				} else {

					color.Println("@{r} El nombre del archivo o carpeta no puede tener más de 20 caracteres")

				}

			} else {
				color.Println("@{r}Path incorrecto, debe iniciar con @{w}/")
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios para la funcion find.")
		}

	} else {
		color.Println("@{r}Se necesita de una sesión activa para ejecutar la función FIND.")
	}

}

//FindAll despliega todas las carpetas y archivos
func FindAll(id string, path string, name string) {

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

				/////////////////////////////////////////// RECORRER Y BUSCAR FILE

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRCUT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDAux := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDAux))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				AnteriorData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				var NombreAnterior [20]byte
				copy(NombreAnterior[:], AVDAux.NombreDir[:])
				//Vamos a comparar Padres e Hijos
				carpetas := strings.Split(path, "/")
				i := 1

				PathCorrecto := true
				for i < len(carpetas)-1 {

					Continuar := true
					//Recorremos el struct y si el apuntador indirecto a punta a otro AVD tambien lo recorreremos en caso que no se encuentre
					//el directorio
					for Continuar {
						//Iteramos en las 6 posiciones del arreglo de subdirectoios (apuntadores)
						for x := 0; x < 6; x++ {
							//Validamos que el apuntador si esté apuntando a algo
							if AVDAux.ApuntadorSubs[x] > 0 {
								//Con el valor del apuntador leemos un struct AVD
								AVDHijo := estructuras.AVD{}
								fileMBR.Seek(int64(AVDAux.ApuntadorSubs[x]+int32(1)), 0)
								HijoData := leerBytes(fileMBR, int(SizeAVD))
								buffer := bytes.NewBuffer(HijoData)
								err = binary.Read(buffer, binary.BigEndian, &AVDHijo)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
								}
								//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
								//si existe el directorio retornamos true y el byte donde está dicho AVD
								var chars [20]byte
								copy(chars[:], carpetas[i])

								if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

									ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									AnteriorData = leerBytes(fileMBR, int(SizeAVD))
									buffer2 = bytes.NewBuffer(AnteriorData)
									err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}
									copy(NombreAnterior[:], AVDAux.NombreDir[:])
									i++
									PathCorrecto = true
									Continuar = false
									break
								}
							}

						}

						if Continuar == false {
							continue
						}

						//Si el directorio no está en el arreglo de apuntadores directos
						//verificamos si el AVD actual apunta hacia otro AVD con otros 6 apuntadores
						if AVDAux.ApuntadorAVD > 0 {

							//Leemos el AVD (que se considera contiguo)
							fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 := bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

						} else {
							//Si ya no apunta a otro AVD y llegamos a esta parte, cancelamos el ciclo FOR
							Continuar = false
							PathCorrecto = false
							break
						}

					}

					if PathCorrecto == false {
						break
					}

				}

				if PathCorrecto {

					if path != "/" {

						if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

							ApuntadorAVD = int32(ApuntadorSiguiente)
							fileMBR.Seek(int64(ApuntadorAVD+1), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 = bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

							if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

								//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
								cadenaConsola = ""
								RecorrerAVDRecursivo(fileMBR, &AVDAux)

							} else {
								PathCorrecto = false
								color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
							}

						} else {
							color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
						}

					} else {

						if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

							//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
							cadenaConsola = ""
							RecorrerAVDRecursivo(fileMBR, &AVDAux)

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

				///////////////////////////////////////////////////

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

				//////////////////////////////////////////

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRCUT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDAux := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDAux))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				AnteriorData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				var NombreAnterior [20]byte
				copy(NombreAnterior[:], AVDAux.NombreDir[:])
				//Vamos a comparar Padres e Hijos
				carpetas := strings.Split(path, "/")
				i := 1

				PathCorrecto := true
				for i < len(carpetas)-1 {

					Continuar := true
					//Recorremos el struct y si el apuntador indirecto a punta a otro AVD tambien lo recorreremos en caso que no se encuentre
					//el directorio
					for Continuar {
						//Iteramos en las 6 posiciones del arreglo de subdirectoios (apuntadores)
						for x := 0; x < 6; x++ {
							//Validamos que el apuntador si esté apuntando a algo
							if AVDAux.ApuntadorSubs[x] > 0 {
								//Con el valor del apuntador leemos un struct AVD
								AVDHijo := estructuras.AVD{}
								fileMBR.Seek(int64(AVDAux.ApuntadorSubs[x]+int32(1)), 0)
								HijoData := leerBytes(fileMBR, int(SizeAVD))
								buffer := bytes.NewBuffer(HijoData)
								err = binary.Read(buffer, binary.BigEndian, &AVDHijo)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
								}
								//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
								//si existe el directorio retornamos true y el byte donde está dicho AVD
								var chars [20]byte
								copy(chars[:], carpetas[i])

								if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

									ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									AnteriorData = leerBytes(fileMBR, int(SizeAVD))
									buffer2 = bytes.NewBuffer(AnteriorData)
									err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}
									copy(NombreAnterior[:], AVDAux.NombreDir[:])
									i++
									PathCorrecto = true
									Continuar = false
									break
								}
							}

						}

						if Continuar == false {
							continue
						}

						//Si el directorio no está en el arreglo de apuntadores directos
						//verificamos si el AVD actual apunta hacia otro AVD con otros 6 apuntadores
						if AVDAux.ApuntadorAVD > 0 {

							//Leemos el AVD (que se considera contiguo)
							fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 := bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

						} else {
							//Si ya no apunta a otro AVD y llegamos a esta parte, cancelamos el ciclo FOR
							Continuar = false
							PathCorrecto = false
							break
						}

					}

					if PathCorrecto == false {
						break
					}

				}

				if PathCorrecto {

					if path != "/" {

						if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

							ApuntadorAVD = int32(ApuntadorSiguiente)
							fileMBR.Seek(int64(ApuntadorAVD+1), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 = bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

							if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

								//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
								cadenaConsola = ""
								RecorrerAVDRecursivo(fileMBR, &AVDAux)

							} else {
								PathCorrecto = false
								color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
							}

						} else {
							color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
						}

					} else {

						if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

							//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
							cadenaConsola = ""
							RecorrerAVDRecursivo(fileMBR, &AVDAux)

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

				///////////////////////////////////////////

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
	}

}

//FindCaso1 despliega todas las carpetas y archivos que cumplen con ?.*
func FindCaso1(id string, path string, name string) {

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

				//////////////////////////////////////////

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRCUT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDAux := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDAux))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				AnteriorData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				var NombreAnterior [20]byte
				copy(NombreAnterior[:], AVDAux.NombreDir[:])
				//Vamos a comparar Padres e Hijos
				carpetas := strings.Split(path, "/")
				i := 1

				PathCorrecto := true
				for i < len(carpetas)-1 {

					Continuar := true
					//Recorremos el struct y si el apuntador indirecto a punta a otro AVD tambien lo recorreremos en caso que no se encuentre
					//el directorio
					for Continuar {
						//Iteramos en las 6 posiciones del arreglo de subdirectoios (apuntadores)
						for x := 0; x < 6; x++ {
							//Validamos que el apuntador si esté apuntando a algo
							if AVDAux.ApuntadorSubs[x] > 0 {
								//Con el valor del apuntador leemos un struct AVD
								AVDHijo := estructuras.AVD{}
								fileMBR.Seek(int64(AVDAux.ApuntadorSubs[x]+int32(1)), 0)
								HijoData := leerBytes(fileMBR, int(SizeAVD))
								buffer := bytes.NewBuffer(HijoData)
								err = binary.Read(buffer, binary.BigEndian, &AVDHijo)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
								}
								//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
								//si existe el directorio retornamos true y el byte donde está dicho AVD
								var chars [20]byte
								copy(chars[:], carpetas[i])

								if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

									ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									AnteriorData = leerBytes(fileMBR, int(SizeAVD))
									buffer2 = bytes.NewBuffer(AnteriorData)
									err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}
									copy(NombreAnterior[:], AVDAux.NombreDir[:])
									i++
									PathCorrecto = true
									Continuar = false
									break
								}
							}

						}

						if Continuar == false {
							continue
						}

						//Si el directorio no está en el arreglo de apuntadores directos
						//verificamos si el AVD actual apunta hacia otro AVD con otros 6 apuntadores
						if AVDAux.ApuntadorAVD > 0 {

							//Leemos el AVD (que se considera contiguo)
							fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 := bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

						} else {
							//Si ya no apunta a otro AVD y llegamos a esta parte, cancelamos el ciclo FOR
							Continuar = false
							PathCorrecto = false
							break
						}

					}

					if PathCorrecto == false {
						break
					}

				}

				if PathCorrecto {

					if path != "/" {

						if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

							ApuntadorAVD = int32(ApuntadorSiguiente)
							fileMBR.Seek(int64(ApuntadorAVD+1), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 = bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

							if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

								//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
								cadenaConsola = ""
								RecorrerAVDRecursivoCaso1(fileMBR, &AVDAux)

							} else {
								PathCorrecto = false
								color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
							}

						} else {
							color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
						}

					} else {

						if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

							//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
							cadenaConsola = ""
							RecorrerAVDRecursivoCaso1(fileMBR, &AVDAux)

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

				////////////////////////////////////////

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

				////////////////////////////////////////////

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRCUT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDAux := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDAux))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				AnteriorData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				var NombreAnterior [20]byte
				copy(NombreAnterior[:], AVDAux.NombreDir[:])
				//Vamos a comparar Padres e Hijos
				carpetas := strings.Split(path, "/")
				i := 1

				PathCorrecto := true
				for i < len(carpetas)-1 {

					Continuar := true
					//Recorremos el struct y si el apuntador indirecto a punta a otro AVD tambien lo recorreremos en caso que no se encuentre
					//el directorio
					for Continuar {
						//Iteramos en las 6 posiciones del arreglo de subdirectoios (apuntadores)
						for x := 0; x < 6; x++ {
							//Validamos que el apuntador si esté apuntando a algo
							if AVDAux.ApuntadorSubs[x] > 0 {
								//Con el valor del apuntador leemos un struct AVD
								AVDHijo := estructuras.AVD{}
								fileMBR.Seek(int64(AVDAux.ApuntadorSubs[x]+int32(1)), 0)
								HijoData := leerBytes(fileMBR, int(SizeAVD))
								buffer := bytes.NewBuffer(HijoData)
								err = binary.Read(buffer, binary.BigEndian, &AVDHijo)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
								}
								//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
								//si existe el directorio retornamos true y el byte donde está dicho AVD
								var chars [20]byte
								copy(chars[:], carpetas[i])

								if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

									ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									AnteriorData = leerBytes(fileMBR, int(SizeAVD))
									buffer2 = bytes.NewBuffer(AnteriorData)
									err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}
									copy(NombreAnterior[:], AVDAux.NombreDir[:])
									i++
									PathCorrecto = true
									Continuar = false
									break
								}
							}

						}

						if Continuar == false {
							continue
						}

						//Si el directorio no está en el arreglo de apuntadores directos
						//verificamos si el AVD actual apunta hacia otro AVD con otros 6 apuntadores
						if AVDAux.ApuntadorAVD > 0 {

							//Leemos el AVD (que se considera contiguo)
							fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 := bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

						} else {
							//Si ya no apunta a otro AVD y llegamos a esta parte, cancelamos el ciclo FOR
							Continuar = false
							PathCorrecto = false
							break
						}

					}

					if PathCorrecto == false {
						break
					}

				}

				if PathCorrecto {

					if path != "/" {

						if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

							ApuntadorAVD = int32(ApuntadorSiguiente)
							fileMBR.Seek(int64(ApuntadorAVD+1), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 = bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

							if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

								//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
								cadenaConsola = ""
								RecorrerAVDRecursivoCaso1(fileMBR, &AVDAux)

							} else {
								PathCorrecto = false
								color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
							}

						} else {
							color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
						}

					} else {

						if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

							//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
							cadenaConsola = ""
							RecorrerAVDRecursivoCaso1(fileMBR, &AVDAux)

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

				/////////////////////////////////////////////

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
	}

}

//FindCaso2 despliega todas las carpetas y archivos que se llaman igual que name
func FindCaso2(id string, path string, name string) {

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

				//////////////////////////////////////////////////////////

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRCUT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDAux := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDAux))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				AnteriorData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				var NombreAnterior [20]byte
				copy(NombreAnterior[:], AVDAux.NombreDir[:])
				//Vamos a comparar Padres e Hijos
				carpetas := strings.Split(path, "/")
				i := 1

				PathCorrecto := true
				for i < len(carpetas)-1 {

					Continuar := true
					//Recorremos el struct y si el apuntador indirecto a punta a otro AVD tambien lo recorreremos en caso que no se encuentre
					//el directorio
					for Continuar {
						//Iteramos en las 6 posiciones del arreglo de subdirectoios (apuntadores)
						for x := 0; x < 6; x++ {
							//Validamos que el apuntador si esté apuntando a algo
							if AVDAux.ApuntadorSubs[x] > 0 {
								//Con el valor del apuntador leemos un struct AVD
								AVDHijo := estructuras.AVD{}
								fileMBR.Seek(int64(AVDAux.ApuntadorSubs[x]+int32(1)), 0)
								HijoData := leerBytes(fileMBR, int(SizeAVD))
								buffer := bytes.NewBuffer(HijoData)
								err = binary.Read(buffer, binary.BigEndian, &AVDHijo)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
								}
								//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
								//si existe el directorio retornamos true y el byte donde está dicho AVD
								var chars [20]byte
								copy(chars[:], carpetas[i])

								if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

									ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									AnteriorData = leerBytes(fileMBR, int(SizeAVD))
									buffer2 = bytes.NewBuffer(AnteriorData)
									err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}
									copy(NombreAnterior[:], AVDAux.NombreDir[:])
									i++
									PathCorrecto = true
									Continuar = false
									break
								}
							}

						}

						if Continuar == false {
							continue
						}

						//Si el directorio no está en el arreglo de apuntadores directos
						//verificamos si el AVD actual apunta hacia otro AVD con otros 6 apuntadores
						if AVDAux.ApuntadorAVD > 0 {

							//Leemos el AVD (que se considera contiguo)
							fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 := bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

						} else {
							//Si ya no apunta a otro AVD y llegamos a esta parte, cancelamos el ciclo FOR
							Continuar = false
							PathCorrecto = false
							break
						}

					}

					if PathCorrecto == false {
						break
					}

				}

				if PathCorrecto {

					if path != "/" {

						if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

							ApuntadorAVD = int32(ApuntadorSiguiente)
							fileMBR.Seek(int64(ApuntadorAVD+1), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 = bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

							if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

								//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
								cadenaConsola = ""
								RecorrerAVDRecursivoCaso2(fileMBR, &AVDAux, name)

							} else {
								PathCorrecto = false
								color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
							}

						} else {
							color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
						}

					} else {

						if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

							//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
							cadenaConsola = ""
							RecorrerAVDRecursivoCaso2(fileMBR, &AVDAux, name)

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

				//////////////////////////////////////////////////////////

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

				//////////////////////////////////////////////

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRCUT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDAux := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDAux))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				AnteriorData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				var NombreAnterior [20]byte
				copy(NombreAnterior[:], AVDAux.NombreDir[:])
				//Vamos a comparar Padres e Hijos
				carpetas := strings.Split(path, "/")
				i := 1

				PathCorrecto := true
				for i < len(carpetas)-1 {

					Continuar := true
					//Recorremos el struct y si el apuntador indirecto a punta a otro AVD tambien lo recorreremos en caso que no se encuentre
					//el directorio
					for Continuar {
						//Iteramos en las 6 posiciones del arreglo de subdirectoios (apuntadores)
						for x := 0; x < 6; x++ {
							//Validamos que el apuntador si esté apuntando a algo
							if AVDAux.ApuntadorSubs[x] > 0 {
								//Con el valor del apuntador leemos un struct AVD
								AVDHijo := estructuras.AVD{}
								fileMBR.Seek(int64(AVDAux.ApuntadorSubs[x]+int32(1)), 0)
								HijoData := leerBytes(fileMBR, int(SizeAVD))
								buffer := bytes.NewBuffer(HijoData)
								err = binary.Read(buffer, binary.BigEndian, &AVDHijo)
								if err != nil {
									fileMBR.Close()
									fmt.Println(err)
									return
								}
								//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
								//si existe el directorio retornamos true y el byte donde está dicho AVD
								var chars [20]byte
								copy(chars[:], carpetas[i])

								if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

									ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									AnteriorData = leerBytes(fileMBR, int(SizeAVD))
									buffer2 = bytes.NewBuffer(AnteriorData)
									err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
									if err != nil {
										fileMBR.Close()
										fmt.Println(err)
										return
									}
									copy(NombreAnterior[:], AVDAux.NombreDir[:])
									i++
									PathCorrecto = true
									Continuar = false
									break
								}
							}

						}

						if Continuar == false {
							continue
						}

						//Si el directorio no está en el arreglo de apuntadores directos
						//verificamos si el AVD actual apunta hacia otro AVD con otros 6 apuntadores
						if AVDAux.ApuntadorAVD > 0 {

							//Leemos el AVD (que se considera contiguo)
							fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 := bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

						} else {
							//Si ya no apunta a otro AVD y llegamos a esta parte, cancelamos el ciclo FOR
							Continuar = false
							PathCorrecto = false
							break
						}

					}

					if PathCorrecto == false {
						break
					}

				}

				if PathCorrecto {

					if path != "/" {

						if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

							ApuntadorAVD = int32(ApuntadorSiguiente)
							fileMBR.Seek(int64(ApuntadorAVD+1), 0)
							AnteriorData = leerBytes(fileMBR, int(SizeAVD))
							buffer2 = bytes.NewBuffer(AnteriorData)
							err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
							if err != nil {
								fileMBR.Close()
								fmt.Println(err)
								return
							}

							if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

								//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
								cadenaConsola = ""
								RecorrerAVDRecursivoCaso2(fileMBR, &AVDAux, name)

							} else {
								PathCorrecto = false
								color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
							}

						} else {
							color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
						}

					} else {

						if sesionRoot || LecturaPropietarioDir(&AVDAux) || LecturaGrupoDir(&AVDAux, id) || LecturaOtrosDir(&AVDAux) {

							//En esta parte podemos comenzar a recorrer el sistema LWH completo e imprimirlo en consola
							cadenaConsola = ""
							RecorrerAVDRecursivoCaso2(fileMBR, &AVDAux, name)

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de lectura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

				////////////////////////////////////////////////

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
	}

}

//LecturaPropietarioDir verifica si un usuario tiene permisos sobre un directorio por ser propietario
func LecturaPropietarioDir(AVDAux *estructuras.AVD) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	//Verificamos si el usuario activo actualmente es el propietario, si no lo es automaticamente returnamos false
	if string(AVDAux.Proper[:]) == string(chars[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if AVDAux.PermisoU == 4 || AVDAux.PermisoU == 5 || AVDAux.PermisoU == 6 || AVDAux.PermisoU == 7 {
			return true
		}
	}

	return false
}

//LecturaGrupoDir verifica si un usuario tiene permisos sobre un directorio por ser parte del grupo
func LecturaGrupoDir(AVDAux *estructuras.AVD, id string) bool {

	var chars [20]byte
	copy(chars[:], idGrupo)

	n := bytes.Index(chars[:], []byte{0})
	if n == -1 {
		n = len(chars)
	}
	GrupoAux := string(chars[:n])

	if GrupoExiste := ExisteGrupo(GrupoAux, id); GrupoExiste {
		//Verificamos si el usuario activo actualmente es parte del grupo, si no lo es automaticamente retornamos false
		if string(AVDAux.Grupo[:]) == string(chars[:]) {
			//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
			if AVDAux.PermisoG == 4 || AVDAux.PermisoG == 5 || AVDAux.PermisoG == 6 || AVDAux.PermisoG == 7 {
				return true
			}
		}

	}

	return false
}

//LecturaOtrosDir verifica si un usuario tiene permisos sobre un directorio por ser de la categoria "Otros"
func LecturaOtrosDir(AVDAux *estructuras.AVD) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	var chars2 [20]byte
	copy(chars2[:], idGrupo)
	//Verificamos si el usuario activo actualmente no es propietario y tampoco parte del grupo, si lo es automaticamente retornamos false
	if string(AVDAux.Proper[:]) != string(chars[:]) && string(AVDAux.Grupo[:]) != string(chars2[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if AVDAux.PermisoO == 4 || AVDAux.PermisoO == 5 || AVDAux.PermisoO == 6 || AVDAux.PermisoO == 7 {
			return true
		}
	}

	return false
}

//RecorrerAVDRecursivo recorre un AVD
func RecorrerAVDRecursivo(file *os.File, AVDAux *estructuras.AVD) {

	color.Printf("@{w}%v|%v\n", cadenaConsola, string(AVDAux.NombreDir[:]))

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

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

			cadenaConsola += " "
			RecorrerAVDRecursivo(file, &AVDHijo)
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
			}

		}
	}

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
	cadenaConsola += " "
	RecorrerDDRecursivo(file, &DDAux)
	if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
		cadenaConsola = cadenaConsola[:last]
	}

	if AVDAux.ApuntadorAVD > 0 {

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

		RecorrerAVDExtRecursivo(file, &AVDExt)
	}

}

//RecorrerAVDExtRecursivo recorre la extensión del AVD
func RecorrerAVDExtRecursivo(file *os.File, AVDAux *estructuras.AVD) {

	//color.Printf("@{w}%v|%v\n", cadenaConsola, string(AVDAux.NombreDir[:]))

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

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

			cadenaConsola += " "
			RecorrerAVDRecursivo(file, &AVDHijo)
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
			}

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

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

		RecorrerAVDExtRecursivo(file, &AVDExt)
	}

}

//RecorrerDDRecursivo recorre el detalle de directorio
func RecorrerDDRecursivo(file *os.File, DDaux *estructuras.DD) {

	for i := 0; i < 5; i++ {

		if DDaux.DDFiles[i].ApuntadorInodo > 0 {
			color.Printf("@{w}%v|%v\n", cadenaConsola, string(DDaux.DDFiles[i].Name[:]))
		}
	}

	if DDaux.ApuntadorDD > 0 {

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

		RecorrerDDRecursivo(file, &DDExt)
	}

}

//RecorrerAVDRecursivoCaso1 recorre un AVD
func RecorrerAVDRecursivoCaso1(file *os.File, AVDAux *estructuras.AVD) {

	color.Printf("@{w}%v|%v\n", cadenaConsola, string(AVDAux.NombreDir[:]))

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

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

			cadenaConsola += " "
			RecorrerAVDRecursivoCaso1(file, &AVDHijo)
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
			}

		}
	}

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
	cadenaConsola += " "
	RecorrerDDRecursivoCaso1(file, &DDAux)
	if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
		cadenaConsola = cadenaConsola[:last]
	}

	if AVDAux.ApuntadorAVD > 0 {

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

		RecorrerAVDExtRecursivoCaso1(file, &AVDExt)
	}

}

//RecorrerAVDExtRecursivoCaso1 recorre la extensión del AVD
func RecorrerAVDExtRecursivoCaso1(file *os.File, AVDAux *estructuras.AVD) {

	//color.Printf("@{w}%v|%v\n", cadenaConsola, string(AVDAux.NombreDir[:]))

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

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

			cadenaConsola += " "
			RecorrerAVDRecursivoCaso1(file, &AVDHijo)
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
			}

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

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

		RecorrerAVDExtRecursivoCaso1(file, &AVDExt)
	}

}

//RecorrerDDRecursivoCaso1 recorre el detalle de directorio
func RecorrerDDRecursivoCaso1(file *os.File, DDaux *estructuras.DD) {

	for i := 0; i < 5; i++ {

		if DDaux.DDFiles[i].ApuntadorInodo > 0 {

			n := bytes.Index(DDaux.DDFiles[i].Name[:], []byte{0})
			if n == -1 {
				n = len(DDaux.DDFiles[i].Name)
			}
			NombreArchivo := string(DDaux.DDFiles[i].Name[:n])

			if matched, _ := regexp.MatchString(`^[^.][.][^.]+$`, NombreArchivo); matched {
				color.Printf("@{w}%v|%v\n", cadenaConsola, NombreArchivo)
			}

		}
	}

	if DDaux.ApuntadorDD > 0 {

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

		RecorrerDDRecursivoCaso1(file, &DDExt)
	}

}

//RecorrerAVDRecursivoCaso2 recorre un AVD
func RecorrerAVDRecursivoCaso2(file *os.File, AVDAux *estructuras.AVD, nombre string) {

	color.Printf("@{w}%v|%v\n", cadenaConsola, string(AVDAux.NombreDir[:]))

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

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

			cadenaConsola += " "
			RecorrerAVDRecursivoCaso2(file, &AVDHijo, nombre)
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
			}

		}
	}

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
	cadenaConsola += " "
	RecorrerDDRecursivoCaso2(file, &DDAux, nombre)
	if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
		cadenaConsola = cadenaConsola[:last]
	}

	if AVDAux.ApuntadorAVD > 0 {

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

		RecorrerAVDExtRecursivoCaso2(file, &AVDExt, nombre)
	}

}

//RecorrerAVDExtRecursivoCaso2 recorre la extensión del AVD
func RecorrerAVDExtRecursivoCaso2(file *os.File, AVDAux *estructuras.AVD, nombre string) {

	//color.Printf("@{w}%v|%v\n", cadenaConsola, string(AVDAux.NombreDir[:]))

	for i := 0; i < 6; i++ {

		if AVDAux.ApuntadorSubs[i] > 0 {

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

			cadenaConsola += " "
			RecorrerAVDRecursivoCaso2(file, &AVDHijo, nombre)
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
			}

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

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

		RecorrerAVDExtRecursivoCaso2(file, &AVDExt, nombre)
	}

}

//RecorrerDDRecursivoCaso2 recorre el detalle de directorio
func RecorrerDDRecursivoCaso2(file *os.File, DDaux *estructuras.DD, nombre string) {

	for i := 0; i < 5; i++ {

		if DDaux.DDFiles[i].ApuntadorInodo > 0 {

			n := bytes.Index(DDaux.DDFiles[i].Name[:], []byte{0})
			if n == -1 {
				n = len(DDaux.DDFiles[i].Name)
			}
			NombreArchivo := string(DDaux.DDFiles[i].Name[:n])

			if NombreArchivo == nombre {
				color.Printf("@{w}%v|%v\n", cadenaConsola, NombreArchivo)
			}

		}
	}

	if DDaux.ApuntadorDD > 0 {

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

		RecorrerDDRecursivoCaso2(file, &DDExt, nombre)
	}

}
