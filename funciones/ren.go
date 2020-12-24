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
	"time"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarRen function
func EjecutarRen(id string, path string, name string) {

	if sesionActiva {

		if path != "" && id != "" && name != "" {

			if strings.HasPrefix(path, "/") {

				if len(name) <= 20 {

					extension := filepath.Ext(path)

					if strings.ToLower(extension) == ".txt" || strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".mia" || strings.ToLower(extension) == ".dsk" || strings.ToLower(extension) == ".sh" {

						if path != "/users.txt" {
							RenameFile(id, path, name)

						} else {
							color.Println("@{r}El sistema no permite cambiar de nombre a archivo @{w}users.txt")
						}

					} else {

						if path != "/" {
							RenameDir(id, path, name)

						} else {
							color.Println("@{r}El sistema no permite cambiar de nombre a carpeta @{w}/")
						}

					}

				} else {

					color.Println("@{r} El nuevo nombre no puede tener más de 20 caracteres")

				}

			} else {
				color.Println("@{r}Path incorrecto, debe iniciar con @{w}/")
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios para la funcion REN.")
		}

	} else {
		color.Println("@{r}Se necesita de una sesión activa para ejecutar la COMANDO REN.")
	}

}

//RenameFile cambia el nombre de un archivo
func RenameFile(id string, path string, name string) {

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

									if sesionRoot || EscrituraPropietarioFile(&InodoAux) || EscrituraGrupoFile(&InodoAux, id) || EscrituraOtrosFile(&InodoAux) {

										if YaExiste := ExisteFile(name, int(ApuntadorAVD), PathAux); !YaExiste {

											copy(NombreAnterior[:], AVDAux.NombreDir[:])
											//Seteamos el nuevo nombre
											var NameChars [20]byte
											copy(NameChars[:], name)
											copy(DDAux.DDFiles[i].Name[:], NameChars[:])

											//Ahora toca reescribir el DD en su posición correspondiente
											fileMBR.Seek(int64(PosicionDD+1), 0)
											ddp := &DDAux
											var binario bytes.Buffer
											binary.Write(&binario, binary.BigEndian, ddp)
											escribirBytes(fileMBR, binario.Bytes())

											//Crear bitacora Ren
											//Creamos la bitacora para la creación de la carpeta
											BitacoraAux := estructuras.Bitacora{}
											//Seteamos el path, en este caso la primera carpeta tiene "/" como path
											var PathChars [300]byte
											PathAux := path
											copy(PathChars[:], PathAux)
											copy(BitacoraAux.Path[:], PathChars[:])
											//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
											var OperacionChars [16]byte
											OperacionAux := "Ren"
											copy(OperacionChars[:], OperacionAux)
											copy(BitacoraAux.Operacion[:], OperacionChars[:])
											//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
											BitacoraAux.Tipo = 0
											//Setemos el contenido
											ContenidoRen := name
											var ContenidoChars [300]byte
											copy(ContenidoChars[:], ContenidoRen)
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

											color.Printf("@{w}El archivo @{w}%v @{w}fue renombrado como @{w}%v @{w}.\n", string(chars[:]), name)

										} else {
											color.Printf("@{r}Ya existe un archivo llamado @{w}%v @{r}en la carpeta @{w}%v.\n", name, string(NombreAnterior[:]))
										}

									} else {
										PathCorrecto = false
										color.Printf("@{r} El usuario @{w}%v @{r}no tiene permisos de escritura sobre el archivo @{w}%v.\n", idSesion, carpetas[len(carpetas)-1])
									}

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

									if sesionRoot || EscrituraPropietarioFile(&InodoAux) || EscrituraGrupoFile(&InodoAux, id) || EscrituraOtrosFile(&InodoAux) {

										if YaExiste := ExisteFile(name, int(ApuntadorAVD), PathAux); !YaExiste {

											copy(NombreAnterior[:], AVDAux.NombreDir[:])
											//Seteamos el nuevo nombre
											var NameChars [20]byte
											copy(NameChars[:], name)
											copy(DDAux.DDFiles[i].Name[:], NameChars[:])

											//Ahora toca reescribir el DD en su posición correspondiente
											fileMBR.Seek(int64(PosicionDD+1), 0)
											ddp := &DDAux
											var binario bytes.Buffer
											binary.Write(&binario, binary.BigEndian, ddp)
											escribirBytes(fileMBR, binario.Bytes())

											//Crear bitacora Ren
											//Creamos la bitacora para la creación de la carpeta
											BitacoraAux := estructuras.Bitacora{}
											//Seteamos el path, en este caso la primera carpeta tiene "/" como path
											var PathChars [300]byte
											PathAux := path
											copy(PathChars[:], PathAux)
											copy(BitacoraAux.Path[:], PathChars[:])
											//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
											var OperacionChars [16]byte
											OperacionAux := "Ren"
											copy(OperacionChars[:], OperacionAux)
											copy(BitacoraAux.Operacion[:], OperacionChars[:])
											//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
											BitacoraAux.Tipo = 0
											//Setemos el contenido
											ContenidoRen := name
											var ContenidoChars [300]byte
											copy(ContenidoChars[:], ContenidoRen)
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

											color.Printf("@{w}El archivo @{w}%v @{w}fue renombrado como @{w}%v @{w}.\n", string(chars[:]), name)

										} else {
											color.Printf("@{r}Ya existe un archivo llamado @{w}%v @{r}en la carpeta @{w}%v.\n", name, string(NombreAnterior[:]))
										}

									} else {
										PathCorrecto = false
										color.Printf("@{r} El usuario @{w}%v @{r}no tiene permisos de escritura sobre el archivo @{w}%v.\n", idSesion, carpetas[len(carpetas)-1])
									}

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

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
	}

}

//RenameDir cambia el nombre de un directorio
func RenameDir(id string, path string, name string) {

	if last := len(path) - 1; last >= 0 && path[last] == '/' {
		path = path[:last]
	}

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

					if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

						ApuntadorCarpeta := int(ApuntadorAVD)

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

						if sesionRoot || EscrituraPropietarioDir(&AVDAux) || EscrituraGrupoDir(&AVDAux, id) || EscrituraOtrosDir(&AVDAux) {

							//Cambiar el nombre a AVD aux y a sus extensiones si las tuviera

							if YaExiste, _ := ExisteSub(name, int(ApuntadorCarpeta), PathAux); !YaExiste {

								copy(NombreAnterior[:], AVDAux.NombreDir[:])

								Continuar := true

								for Continuar {

									//Seteamos el nuevo nombre
									var NameChars [20]byte
									copy(NameChars[:], name)
									copy(AVDAux.NombreDir[:], NameChars[:])

									//Ahora toca reescribir el DD en su posición correspondiente
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									avdp := &AVDAux
									var binario bytes.Buffer
									binary.Write(&binario, binary.BigEndian, avdp)
									escribirBytes(fileMBR, binario.Bytes())

									if AVDAux.ApuntadorAVD > 0 {
										ApuntadorAVD = int32(AVDAux.ApuntadorAVD)
										//Leemos el AVD (que se considera contiguo)
										fileMBR.Seek(int64(ApuntadorAVD+int32(1)), 0)
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

								//Crear bitacora Ren
								//Creamos la bitacora para la creación de la carpeta
								BitacoraAux := estructuras.Bitacora{}
								//Seteamos el path, en este caso la primera carpeta tiene "/" como path
								var PathChars [300]byte
								PathAux := path
								copy(PathChars[:], PathAux)
								copy(BitacoraAux.Path[:], PathChars[:])
								//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
								var OperacionChars [16]byte
								OperacionAux := "Ren"
								copy(OperacionChars[:], OperacionAux)
								copy(BitacoraAux.Operacion[:], OperacionChars[:])
								//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
								BitacoraAux.Tipo = 1
								//Setemos el contenido
								ContenidoRen := name
								var ContenidoChars [300]byte
								copy(ContenidoChars[:], ContenidoRen)
								copy(BitacoraAux.Contenido[:], ContenidoChars[:])
								BitacoraAux.Size = -1
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

								color.Printf("@{w}La carpeta @{w}%v @{w}fue renombrada como @{w}%v @{w}exitosamente.\n", string(NombreAnterior[:]), name)

							} else {
								color.Printf("@{r}Ya existe una carpeta llamada @{w}%v @{r}en la carpeta @{w}%v.\n", name, string(NombreAnterior[:]))
							}

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de escritura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}

					} else {
						color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
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

					if YaExiste, ApuntadorSiguiente := ExisteSub(carpetas[len(carpetas)-1], int(ApuntadorAVD), PathAux); YaExiste {

						ApuntadorCarpeta := int(ApuntadorAVD)

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

						if sesionRoot || EscrituraPropietarioDir(&AVDAux) || EscrituraGrupoDir(&AVDAux, id) || EscrituraOtrosDir(&AVDAux) {

							//Cambiar el nombre a AVD aux y a sus extensiones si las tuviera

							if YaExiste, _ := ExisteSub(name, int(ApuntadorCarpeta), PathAux); !YaExiste {

								copy(NombreAnterior[:], AVDAux.NombreDir[:])

								Continuar := true

								for Continuar {

									//Seteamos el nuevo nombre
									var NameChars [20]byte
									copy(NameChars[:], name)
									copy(AVDAux.NombreDir[:], NameChars[:])

									//Ahora toca reescribir el DD en su posición correspondiente
									fileMBR.Seek(int64(ApuntadorAVD+1), 0)
									avdp := &AVDAux
									var binario bytes.Buffer
									binary.Write(&binario, binary.BigEndian, avdp)
									escribirBytes(fileMBR, binario.Bytes())

									if AVDAux.ApuntadorAVD > 0 {
										ApuntadorAVD = int32(AVDAux.ApuntadorAVD)
										//Leemos el AVD (que se considera contiguo)
										fileMBR.Seek(int64(ApuntadorAVD+int32(1)), 0)
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

								//Crear bitacora Ren
								//Creamos la bitacora para la creación de la carpeta
								BitacoraAux := estructuras.Bitacora{}
								//Seteamos el path, en este caso la primera carpeta tiene "/" como path
								var PathChars [300]byte
								PathAux := path
								copy(PathChars[:], PathAux)
								copy(BitacoraAux.Path[:], PathChars[:])
								//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
								var OperacionChars [16]byte
								OperacionAux := "Ren"
								copy(OperacionChars[:], OperacionAux)
								copy(BitacoraAux.Operacion[:], OperacionChars[:])
								//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
								BitacoraAux.Tipo = 1
								//Setemos el contenido
								ContenidoRen := name
								var ContenidoChars [300]byte
								copy(ContenidoChars[:], ContenidoRen)
								copy(BitacoraAux.Contenido[:], ContenidoChars[:])
								BitacoraAux.Size = -1
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

								color.Printf("@{w}La carpeta @{w}%v @{w}fue renombrada como @{w}%v @{w}exitosamente.\n", string(NombreAnterior[:]), name)

							} else {
								color.Printf("@{r}Ya existe una carpeta llamada @{w}%v @{r}en la carpeta @{w}%v.\n", name, string(NombreAnterior[:]))
							}

						} else {
							PathCorrecto = false
							color.Printf("@{r} El usuario @{w}%v @{w}no tiene permisos de escritura en la carpeta @{w}%v.\n", idSesion, string(NombreAnterior[:]))
						}

					} else {
						color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
					}

				} else {
					color.Println("@{r} Error, una o más carpetas padre no existen.")

				}

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
	}

}
