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

//EjecutarCat function
func EjecutarCat(id string, lista *[]string) {

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

				////////ITERAMOS EN CADA RUTA ENVIADA EN LA LISTA

				for _, file := range *lista {

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

					//Vamos a comparar Padres e Hijos
					carpetas := strings.Split(file, "/")
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
										fmt.Println("")
										color.Printf("@{w}%s: ", carpetas[len(carpetas)-1])

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

										if sesionRoot || LecturaPropietarioFile(&InodoAux) || LecturaGrupoFile(&InodoAux, id) || LecturaOtrosFile(&InodoAux) {

											Continuar2 := true

											for Continuar2 {

												for i := 0; i < 4; i++ {

													if InodoAux.ApuntadoresBloques[i] > 0 {

														//Con el valor del apuntador leemos un struct Bloque
														BloqueAux := estructuras.BloqueDatos{}
														fileMBR.Seek(int64(InodoAux.ApuntadoresBloques[i]+int32(1)), 0)
														SizeBloque := int(unsafe.Sizeof(BloqueAux))
														BloqueData := leerBytes(fileMBR, int(SizeBloque))
														buffer := bytes.NewBuffer(BloqueData)
														err := binary.Read(buffer, binary.BigEndian, &BloqueAux)
														if err != nil {
															fmt.Println(err)
															return

														}

														color.Printf("@{w}%s", string(BloqueAux.Data[:]))

													}

												}

												if InodoAux.ApuntadorIndirecto > 0 {

													//Con el valor del apuntador leemos un struct Inodo
													InodoExt := estructuras.Inodo{}
													fileMBR.Seek(int64(InodoAux.ApuntadorIndirecto+int32(1)), 0)
													SizeInodo := int(unsafe.Sizeof(InodoExt))
													ExtData := leerBytes(fileMBR, int(SizeInodo))
													buffer := bytes.NewBuffer(ExtData)
													err := binary.Read(buffer, binary.BigEndian, &InodoAux)
													if err != nil {
														fmt.Println(err)
														return

													}

												} else {
													Continuar2 = false
												}

												if Continuar2 == false {
													break
												}

											}

										} else {
											PathCorrecto = false
											color.Printf("@{r} El usuario @{w}%v @{r}no tiene permisos de lectura sobre el archivo @{w}%v.\n", idSesion, carpetas[len(carpetas)-1])
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
								fileMBR.Seek(int64(DDAux.ApuntadorDD+int32(1)), 0)
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
						color.Println("@{r} Error, una o más carpetas no existen.")

					}
				}
				fmt.Println("")
				/////////////////////////////////////////// FIN DE BUSQUEDA DEL FILE

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

				////////ITERAMOS EN CADA RUTA ENVIADA EN LA LISTA

				for _, file := range *lista {

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

					//Vamos a comparar Padres e Hijos
					carpetas := strings.Split(file, "/")
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
										fmt.Println("")
										color.Printf("@{w}%s: ", carpetas[len(carpetas)-1])

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

										if sesionRoot || LecturaPropietarioFile(&InodoAux) || LecturaGrupoFile(&InodoAux, id) || LecturaOtrosFile(&InodoAux) {

											Continuar2 := true

											for Continuar2 {

												for i := 0; i < 4; i++ {

													if InodoAux.ApuntadoresBloques[i] > 0 {

														//Con el valor del apuntador leemos un struct Bloque
														BloqueAux := estructuras.BloqueDatos{}
														fileMBR.Seek(int64(InodoAux.ApuntadoresBloques[i]+int32(1)), 0)
														SizeBloque := int(unsafe.Sizeof(BloqueAux))
														BloqueData := leerBytes(fileMBR, int(SizeBloque))
														buffer := bytes.NewBuffer(BloqueData)
														err := binary.Read(buffer, binary.BigEndian, &BloqueAux)
														if err != nil {
															fmt.Println(err)
															return

														}

														color.Printf("@{w}%s", string(BloqueAux.Data[:]))

													}

												}

												if InodoAux.ApuntadorIndirecto > 0 {

													//Con el valor del apuntador leemos un struct Inodo
													InodoExt := estructuras.Inodo{}
													fileMBR.Seek(int64(InodoAux.ApuntadorIndirecto+int32(1)), 0)
													SizeInodo := int(unsafe.Sizeof(InodoExt))
													ExtData := leerBytes(fileMBR, int(SizeInodo))
													buffer := bytes.NewBuffer(ExtData)
													err := binary.Read(buffer, binary.BigEndian, &InodoAux)
													if err != nil {
														fmt.Println(err)
														return

													}

												} else {
													Continuar2 = false
												}

											}

										} else {
											PathCorrecto = false
											color.Printf("@{r} El usuario @{w}%v @{r}no tiene permisos de lectura sobre el archivo @{w}%v.\n", idSesion, carpetas[len(carpetas)-1])
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
								fileMBR.Seek(int64(DDAux.ApuntadorDD+int32(1)), 0)
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
						color.Println("@{r} Error, una o más carpetas no existen.")

					}
				}
				fmt.Println("")
				/////////////////////////////////////////// FIN DE BUSQUEDA DEL FILE

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Printf("@{r}No hay ninguna partición montada con el id: %v\n", id)
	}

}

//LecturaPropietarioFile verifica si un usuario tiene permisos sobre un directorio por ser propietario
func LecturaPropietarioFile(InodoAux *estructuras.Inodo) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	//Verificamos si el usuario activo actualmente es el propietario, si no lo es automaticamente returnamos false
	if string(InodoAux.Proper[:]) == string(chars[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if InodoAux.PermisoU == 4 || InodoAux.PermisoU == 5 || InodoAux.PermisoU == 6 || InodoAux.PermisoU == 7 {
			return true
		}
	}

	return false
}

//LecturaGrupoFile verifica si un usuario tiene permisos sobre un directorio por ser parte del grupo
func LecturaGrupoFile(InodoAux *estructuras.Inodo, id string) bool {

	var chars [20]byte
	copy(chars[:], idGrupo)

	n := bytes.Index(chars[:], []byte{0})
	if n == -1 {
		n = len(chars)
	}
	GrupoAux := string(chars[:n])

	if GrupoExiste := ExisteGrupo(GrupoAux, id); GrupoExiste {
		//Verificamos si el usuario activo actualmente es parte del grupo, si no lo es automaticamente retornamos false
		if string(InodoAux.Grupo[:]) == string(chars[:]) {
			//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
			if InodoAux.PermisoG == 4 || InodoAux.PermisoG == 5 || InodoAux.PermisoG == 6 || InodoAux.PermisoG == 7 {
				return true
			}
		}
	}

	return false
}

//LecturaOtrosFile verifica si un usuario tiene permisos sobre un directorio por ser de la categoria "Otros"
func LecturaOtrosFile(InodoAux *estructuras.Inodo) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	var chars2 [20]byte
	copy(chars2[:], idGrupo)
	//Verificamos si el usuario activo actualmente no es propietario y tampoco parte del grupo, si lo es automaticamente retornamos false
	if string(InodoAux.Proper[:]) != string(chars[:]) && string(InodoAux.Grupo[:]) != string(chars2[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if InodoAux.PermisoO == 4 || InodoAux.PermisoO == 5 || InodoAux.PermisoO == 6 || InodoAux.PermisoO == 7 {
			return true
		}
	}

	return false
}

//EscrituraPropietarioFile verifica si un usuario tiene permisos sobre un directorio por ser propietario
func EscrituraPropietarioFile(InodoAux *estructuras.Inodo) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	//Verificamos si el usuario activo actualmente es el propietario, si no lo es automaticamente returnamos false
	if string(InodoAux.Proper[:]) == string(chars[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if InodoAux.PermisoU == 2 || InodoAux.PermisoU == 3 || InodoAux.PermisoU == 6 || InodoAux.PermisoU == 7 {
			return true
		}
	}

	return false
}

//EscrituraGrupoFile verifica si un usuario tiene permisos sobre un directorio por ser parte del grupo
func EscrituraGrupoFile(InodoAux *estructuras.Inodo, id string) bool {

	var chars [20]byte
	copy(chars[:], idGrupo)

	n := bytes.Index(chars[:], []byte{0})
	if n == -1 {
		n = len(chars)
	}
	GrupoAux := string(chars[:n])

	if GrupoExiste := ExisteGrupo(GrupoAux, id); GrupoExiste {
		//Verificamos si el usuario activo actualmente es parte del grupo, si no lo es automaticamente retornamos false
		if string(InodoAux.Grupo[:]) == string(chars[:]) {
			//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
			if InodoAux.PermisoG == 2 || InodoAux.PermisoG == 3 || InodoAux.PermisoG == 6 || InodoAux.PermisoG == 7 {
				return true
			}
		}

	}

	return false
}

//EscrituraOtrosFile verifica si un usuario tiene permisos sobre un directorio por ser de la categoria "Otros"
func EscrituraOtrosFile(InodoAux *estructuras.Inodo) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	var chars2 [20]byte
	copy(chars2[:], idGrupo)
	//Verificamos si el usuario activo actualmente no es propietario y tampoco parte del grupo, si lo es automaticamente retornamos false
	if string(InodoAux.Proper[:]) != string(chars[:]) && string(InodoAux.Grupo[:]) != string(chars2[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if InodoAux.PermisoO == 2 || InodoAux.PermisoO == 3 || InodoAux.PermisoO == 6 || InodoAux.PermisoO == 7 {
			return true
		}
	}

	return false
}

//LecturaYEscrituraPropietarioFile verifica si un usuario tiene permisos sobre un directorio por ser propietario
func LecturaYEscrituraPropietarioFile(InodoAux *estructuras.Inodo) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	//Verificamos si el usuario activo actualmente es el propietario, si no lo es automaticamente returnamos false
	if string(InodoAux.Proper[:]) == string(chars[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if InodoAux.PermisoU == 6 || InodoAux.PermisoU == 7 {
			return true
		}
	}

	return false
}

//LecturaYEscrituraGrupoFile verifica si un usuario tiene permisos sobre un directorio por ser parte del grupo
func LecturaYEscrituraGrupoFile(InodoAux *estructuras.Inodo, id string) bool {

	var chars [20]byte
	copy(chars[:], idGrupo)

	n := bytes.Index(chars[:], []byte{0})
	if n == -1 {
		n = len(chars)
	}
	GrupoAux := string(chars[:n])

	if GrupoExiste := ExisteGrupo(GrupoAux, id); GrupoExiste {
		//Verificamos si el usuario activo actualmente es parte del grupo, si no lo es automaticamente retornamos false
		if string(InodoAux.Grupo[:]) == string(chars[:]) {
			//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
			if InodoAux.PermisoG == 6 || InodoAux.PermisoG == 7 {
				return true
			}
		}

	}

	return false
}

//LecturaYEscrituraOtrosFile verifica si un usuario tiene permisos sobre un directorio por ser de la categoria "Otros"
func LecturaYEscrituraOtrosFile(InodoAux *estructuras.Inodo) bool {

	var chars [20]byte
	copy(chars[:], idSesion)
	var chars2 [20]byte
	copy(chars2[:], idGrupo)
	//Verificamos si el usuario activo actualmente no es propietario y tampoco parte del grupo, si lo es automaticamente retornamos false
	if string(InodoAux.Proper[:]) != string(chars[:]) && string(InodoAux.Grupo[:]) != string(chars2[:]) {
		//Si es el propietario verificamos que el directorio tenga permisos de escritura en el parámeto U
		if InodoAux.PermisoO == 6 || InodoAux.PermisoO == 7 {
			return true
		}
	}

	return false
}
