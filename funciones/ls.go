package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarLS crea el reporte LS
func EjecutarLS(path string, id string) {

	if strings.HasPrefix(path, "/") {

		extension := filepath.Ext(path)

		if strings.ToLower(extension) == ".txt" || strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".mia" || strings.ToLower(extension) == ".dsk" || strings.ToLower(extension) == ".sh" {
			LSFile(path, id)
		} else {
			if path != "/" { // si no es root quita slash al final
				if last := len(path) - 1; last >= 0 && path[last] == '/' {
					path = path[:last]
				}
			}
			LSDir(path, id)

		}

	} else {
		color.Println("@{r}Path incorrecto, debe iniciar con @{w}/")
	}

}

//LSFile crea el reporte LS para un archivo
func LSFile(path string, id string) {

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

				//AHORA DEBEMOS LEER EL DETALLE DIRECTORIO DE DICHO AVD
				DDAux := estructuras.DD{}
				PosicionDD := AVDAux.ApuntadorDD
				SizeDD := int(unsafe.Sizeof(DDAux))
				fileMBR.Seek(int64(PosicionDD+1), 0)
				DDData := leerBytes(fileMBR, int(SizeDD))
				bufferDD := bytes.NewBuffer(DDData)
				err = binary.Read(bufferDD, binary.BigEndian, &DDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}
				Continuar := true
				//Recorremos el struct DD, y si el apuntador indirecto a apunta a otro DD tambien lo recorremos
				//en caso que no se encuentre el archivo
				for Continuar {
					//Iteramos en las 5 posiciones del arreglo de archivos que tiene el DD
					for i := 0; i < 5; i++ {
						//Validamos que el apuntador al inodo si esté apuntando a algo
						if DDAux.DDFiles[i].ApuntadorInodo > 0 {
							//Comparamos el nombre del archivo con el nombre del archivo que queremos verificar si existe
							//si existe el archivo retornamos true
							var chars [20]byte
							copy(chars[:], carpetas[len(carpetas)-1])

							if string(DDAux.DDFiles[i].Name[:]) == string(chars[:]) {

								//Con el valor del apuntador leemos un struct Inodo
								InodoAux := estructuras.Inodo{}
								fileMBR.Seek(int64(DDAux.DDFiles[i].ApuntadorInodo+int32(1)), 0)
								SizeInodo := int(unsafe.Sizeof(InodoAux))
								InodoData := leerBytes(fileMBR, int(SizeInodo))
								buffer := bytes.NewBuffer(InodoData)
								err := binary.Read(buffer, binary.BigEndian, &InodoAux)
								if err != nil {
									fmt.Println(err)
									return
								}
								cadenaConsola = ""
								fmt.Println("")
								LsInodoRecursivo(fileMBR, &InodoAux, string(DDAux.DDFiles[i].FechaCreacion[:]), string(DDAux.DDFiles[i].Name[:]))
								fmt.Println("")
								color.Print("@{w}El reporte@{w} LS @{w}fue creado con éxito\n")
								Continuar = false
								break

							}

						}

					}

					if Continuar == false {
						break
					}

					//Si el archivo no está en el arreglo de archivos
					//verificamos si el DD actual apunta hacia otro DD

					if DDAux.ApuntadorDD > 0 {

						//Leemos el DD (que se considera contiguo)
						PosicionDD = DDAux.ApuntadorDD
						fileMBR.Seek(int64(PosicionDD+int32(1)), 0)
						DDData = leerBytes(fileMBR, int(SizeDD))
						bufferDD = bytes.NewBuffer(DDData)
						err = binary.Read(bufferDD, binary.BigEndian, &DDAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return
						}

					} else {
						//Si ya no apunta a otro DD y llegamos a esta parte, cancelamos el ciclo FOR
						Continuar = false
						color.Println("@{r} El archivo no existe.")
						break
					}
				}

			} else {
				color.Println("@{r} Error, una o más carpetas padre no existen.")

			}

			////////////////////////////////////////////////////////////////////////

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

				//AHORA DEBEMOS LEER EL DETALLE DIRECTORIO DE DICHO AVD
				DDAux := estructuras.DD{}
				PosicionDD := AVDAux.ApuntadorDD
				SizeDD := int(unsafe.Sizeof(DDAux))
				fileMBR.Seek(int64(PosicionDD+1), 0)
				DDData := leerBytes(fileMBR, int(SizeDD))
				bufferDD := bytes.NewBuffer(DDData)
				err = binary.Read(bufferDD, binary.BigEndian, &DDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}
				Continuar := true
				//Recorremos el struct DD, y si el apuntador indirecto a apunta a otro DD tambien lo recorremos
				//en caso que no se encuentre el archivo
				for Continuar {
					//Iteramos en las 5 posiciones del arreglo de archivos que tiene el DD
					for i := 0; i < 5; i++ {
						//Validamos que el apuntador al inodo si esté apuntando a algo
						if DDAux.DDFiles[i].ApuntadorInodo > 0 {
							//Comparamos el nombre del archivo con el nombre del archivo que queremos verificar si existe
							//si existe el archivo retornamos true
							var chars [20]byte
							copy(chars[:], carpetas[len(carpetas)-1])

							if string(DDAux.DDFiles[i].Name[:]) == string(chars[:]) {

								//Con el valor del apuntador leemos un struct Inodo
								InodoAux := estructuras.Inodo{}
								fileMBR.Seek(int64(DDAux.DDFiles[i].ApuntadorInodo+int32(1)), 0)
								SizeInodo := int(unsafe.Sizeof(InodoAux))
								InodoData := leerBytes(fileMBR, int(SizeInodo))
								buffer := bytes.NewBuffer(InodoData)
								err := binary.Read(buffer, binary.BigEndian, &InodoAux)
								if err != nil {
									fmt.Println(err)
									return
								}
								cadenaConsola = ""
								fmt.Println("")
								LsInodoRecursivo(fileMBR, &InodoAux, string(DDAux.DDFiles[i].FechaCreacion[:]), string(DDAux.DDFiles[i].Name[:]))
								fmt.Println("")
								color.Print("@{w}El reporte@{w} LS @{w}fue creado con éxito\n")
								Continuar = false
								break

							}

						}

					}

					if Continuar == false {
						break
					}

					//Si el archivo no está en el arreglo de archivos
					//verificamos si el DD actual apunta hacia otro DD

					if DDAux.ApuntadorDD > 0 {

						//Leemos el DD (que se considera contiguo)
						PosicionDD = DDAux.ApuntadorDD
						fileMBR.Seek(int64(PosicionDD+int32(1)), 0)
						DDData = leerBytes(fileMBR, int(SizeDD))
						bufferDD = bytes.NewBuffer(DDData)
						err = binary.Read(bufferDD, binary.BigEndian, &DDAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return
						}

					} else {
						//Si ya no apunta a otro DD y llegamos a esta parte, cancelamos el ciclo FOR
						Continuar = false
						color.Println("@{r} El archivo no existe.")
						break
					}
				}

			} else {
				color.Println("@{r} Error, una o más carpetas padre no existen.")

			}

			////////////////////////////////////////////////////////////////////////

		} else {
			color.Println("@{r} La partición indicada no ha sido formateada.")
		}

		fileMBR.Close()

	}

}

//LSDir crea el reporte LS para un directorio
func LSDir(path string, id string) {

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

						cadenaConsola = ""
						fmt.Println("")
						LsAVDRecursivo(fileMBR, &AVDAux)
						fmt.Println("")
						color.Print("@{w}El reporte@{w} LS @{w}fue creado con éxito\n")

					} else {
						color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
					}

				} else {

					cadenaConsola = ""
					LsAVDRecursivo(fileMBR, &AVDAux)
					color.Print("@{w}El reporte@{w} LS @{w}fue creado con éxito\n")
				}

			} else {
				color.Println("@{r} Error, una o más carpetas padre no existen.")

			}

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

						cadenaConsola = ""
						fmt.Println("")
						LsAVDRecursivo(fileMBR, &AVDAux)
						fmt.Println("")
						color.Print("@{w}El reporte@{w} LS @{w}fue creado con éxito\n")

					} else {
						color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
					}

				} else {

					cadenaConsola = ""
					LsAVDRecursivo(fileMBR, &AVDAux)
					color.Print("@{w}El reporte@{w} LS @{w}fue creado con éxito\n")
				}

			} else {
				color.Println("@{r} Error, una o más carpetas padre no existen.")

			}

		} else {
			color.Println("@{r} La partición indicada no ha sido formateada.")
		}

		fileMBR.Close()

	}

}

//LsAVDRecursivo recorre un AVD
func LsAVDRecursivo(file *os.File, AVDAux *estructuras.AVD) {

	color.Printf("@{w}%v|@{w}Permisos: @{w}%v%v%v @{w}Propietario: @{w}%v @{w}Grupo: @{w}%v @{w}Fecha: @{w}%v @{w}Nombre: @{w}%v\n", cadenaConsola, AVDAux.PermisoU, AVDAux.PermisoG, AVDAux.PermisoO, string(AVDAux.Proper[:]), string(AVDAux.Grupo[:]), string(AVDAux.FechaCreacion[:]), string(AVDAux.NombreDir[:]))

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
			LsAVDRecursivo(file, &AVDHijo)
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

	LsDDRecursivo(file, &DDAux)

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

		LsAVDExtRecursivo(file, &AVDExt)

	}

}

//LsAVDExtRecursivo recorre la extensión del AVD
func LsAVDExtRecursivo(file *os.File, AVDAux *estructuras.AVD) {

	//color.Printf("@{w}%v|@{w}Permisos: @{w}%v%v%v @{w}Propietario: @{w}%v @{w}Grupo: @{w}%v @{w}Fecha: @{w}%v @{w}Nombre: @{w}%v\n", cadenaConsola, AVDAux.PermisoU, AVDAux.PermisoG, AVDAux.PermisoO, string(AVDAux.Proper[:]), string(AVDAux.Grupo[:]), string(AVDAux.FechaCreacion[:]), string(AVDAux.NombreDir[:]))

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
			LsAVDRecursivo(file, &AVDHijo)
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

		LsAVDExtRecursivo(file, &AVDExt)
	}

}

//LsDDRecursivo recorre el detalle de directorio
func LsDDRecursivo(file *os.File, DDaux *estructuras.DD) {

	for i := 0; i < 5; i++ {

		if DDaux.DDFiles[i].ApuntadorInodo > 0 {

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
			cadenaConsola += " "
			LsInodoRecursivo(file, &InodoAux, string(DDaux.DDFiles[i].FechaCreacion[:]), string(DDaux.DDFiles[i].Name[:]))
			if last := len(cadenaConsola) - 1; last >= 0 && cadenaConsola[last] == ' ' {
				cadenaConsola = cadenaConsola[:last]
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

		LsDDRecursivo(file, &DDExt)
	}

}

//LsInodoRecursivo recorre el inodo
func LsInodoRecursivo(file *os.File, InodoAux *estructuras.Inodo, fecha string, nombre string) {

	color.Printf("@{w}%v|@{w}Permisos: @{w}%v%v%v @{w}Propietario: @{w}%v @{w}Grupo: @{w}%v @{w}Fecha: @{w}%v @{w}Nombre: @{w}%v\n", cadenaConsola, InodoAux.PermisoU, InodoAux.PermisoG, InodoAux.PermisoO, string(InodoAux.Proper[:]), string(InodoAux.Grupo[:]), fecha, nombre)

}
