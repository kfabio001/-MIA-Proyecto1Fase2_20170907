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

//EjecutarEdit function
func EjecutarEdit(id string, path string, size string, cont string) {

	if sesionActiva {

		if path != "" && id != "" {

			if strings.HasPrefix(path, "/") {

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

												if sesionRoot || LecturaYEscrituraPropietarioFile(&InodoAux) || LecturaYEscrituraGrupoFile(&InodoAux, id) || LecturaYEscrituraOtrosFile(&InodoAux) {

													if EsCorrecto, Fsize := SizeCorrecto(size); EsCorrecto {

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
																InodoData := leerBytes(fileMBR, SizeInodo)
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
														//EN ESTA PARTE CADENACONTENIDO YA TIENE EL CONTENIDO
														contLeido := string(Contenido[:ContenidoSize])

														if cont != "" {
															contLeido = cont
														}

														//Antes de guardar el nuevo contenido debemos setear a cero los apuntadores y las posiciones en el bitmap
														//LEEMOS EL PRIMER INODO QUE ES EL INODO DEL ARCHIVO USERS.TXT
														InodoPointer := DDAux.DDFiles[i].ApuntadorInodo
														fileMBR.Seek(int64(InodoPointer+1), 0)
														InodoAux = estructuras.Inodo{}
														InodoSize := int(unsafe.Sizeof(InodoAux))
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
																	InodoAux.ApuntadoresBloques[i] = 0
																}
															}

															//Ahora toca escribir el ultimo strcut Inodo creado en su posición correspondiente
															fileMBR.Seek(int64(InodoPointer+1), 0)
															inodop := &InodoAux
															var binario bytes.Buffer
															binary.Write(&binario, binary.BigEndian, inodop)
															escribirBytes(fileMBR, binario.Bytes())

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

														InodoPointer = DDAux.DDFiles[i].ApuntadorInodo
														fileMBR.Seek(int64(InodoPointer+1), 0)
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

														contenido := ""
														FileSize := 0

														//Ahora toca escribir el contenido en los bloques de datos
														//y enlazarlos con newInodo, cabe recalcar que se debe validar
														//la cantidad de bloques necesarios, para crear un nuevo inodo de ser necesario

														if size == "" && contLeido == "" { //Si no vienen nungun parámetro ni size ni cont, el tamaño de 0

															//Si el size es cero, el archivo no contiene datos, por lo tanto no se le asigna ningun bloque
															//y podemos proceder a guardarlo inmediatamente

															InodoAux.FileSize = 0
															InodoAux.NumeroBloques = 0
															InodoAux.ApuntadorIndirecto = 0

														} else {

															if Fsize == 0 && size != "" {

																//Si el size es cero, el archivo no contiene datos, por lo tanto no se le asigna ningun bloque
																//y podemos proceder a guardarlo inmediatamente

																InodoAux.FileSize = 0
																InodoAux.NumeroBloques = 0
																InodoAux.ApuntadorIndirecto = 0

															} else if Fsize == 0 && size == "" && contLeido == "" {

																//Si el size es cero, el archivo no contiene datos, por lo tanto no se le asigna ningun bloque
																//y podemos proceder a guardarlo inmediatamente

																InodoAux.FileSize = 0
																InodoAux.NumeroBloques = 0
																InodoAux.ApuntadorIndirecto = 0

															} else {

																if contLeido == "" { //No hay contenido (setear el abcdario)

																	contenido = getNTimesabc(Fsize)
																	FileSize = Fsize

																} else if size == "" && contLeido != "" {

																	contenido = contLeido
																	FileSize = len(contenido)

																} else {

																	if Fsize > len(contLeido) { //size es mayor que el tamaño de cont, completar cont con el abcdario hasta llegar a size

																		FileSize = Fsize
																		contenido = contLeido        //contenido se le concatena el contenido enviado como parametro, que tiene un tamaño menor a size
																		r := Fsize - len(contLeido)  //calculamos cuantos caraceres hay que agregarle a contenido para cumplir con el size
																		contenido += getNTimesabc(r) //a contenido le enviamos los caraćteres del abcdario necesarios para llegar al size

																	} else if Fsize == len(contLeido) { //si size es igual al tamaño de cont, llenamos los bloques con cont

																		FileSize = Fsize
																		contenido = contLeido

																	} else if Fsize < len(contLeido) { //si size es menor que el tamaño de cont, cortamos cont hasta el tamaño de size

																		FileSize = Fsize
																		contenido = contLeido[:Fsize]

																	}

																}

															}

														}

														//Dividimos la cantidad de caracteres que tiene contenido dentro de 25
														//si el resultado es un numero decimal lo aproximamos al entero más cercano a la derecha
														//esto para saber cuando bloques de datos vamos a usar
														//Ejemplo: Si tenemos 52 caracteres, al dividirlo dentro de 25 obtenemos 2.08
														//ese 0.08 nos obliga a tomar un 3er bloque, entonces la funcion Roundf, aproximaria 2.08 a 3
														CadenaContenido := contenido
														FileSize = len(CadenaContenido)
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
																fileMBR.Seek(int64(InodoPointer+1), 0)
																inodop := &InodoAux
																var binario bytes.Buffer
																binary.Write(&binario, binary.BigEndian, inodop)
																escribirBytes(fileMBR, binario.Bytes())

																//PermisosAux
																valorU := InodoAux.PermisoU
																valorG := InodoAux.PermisoG
																valorO := InodoAux.PermisoO

																var ProperChars [16]byte
																copy(ProperChars[:], idSesion)
																//Setemos el grupo a la bitacora
																var GrupoChars [16]byte
																copy(GrupoChars[:], idGrupo)

																//Ahora creamos el nuevo inodo
																InodoAux = estructuras.Inodo{}
																InodoPointer = int32(Inodo2Pos)

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
																InodoAux.PermisoU = valorU
																InodoAux.PermisoG = valorG
																InodoAux.PermisoO = valorO
																copy(InodoAux.Proper[:], ProperChars[:])
																copy(InodoAux.Grupo[:], GrupoChars[:])

																//Seteamos SB
																SB1.FreeInodos = SB1.FreeInodos - 1

																indx = 0
															}
														}

														InodoAux.FileSize = int32(FileSize)
														InodoAux.NumeroBloques = CantidadBloques

														//Ahora toca escribir el ultimo strcut Inodo creado en su posición correspondiente
														fileMBR.Seek(int64(InodoPointer+1), 0)
														inodop := &InodoAux
														var binario bytes.Buffer
														binary.Write(&binario, binary.BigEndian, inodop)
														escribirBytes(fileMBR, binario.Bytes())

														//Crear bitacora MKUSR
														//Creamos la bitacora para la creación de la carpeta
														BitacoraAux := estructuras.Bitacora{}
														//Seteamos el path, en este caso la primera carpeta tiene "/" como path
														var PathChars [300]byte
														PathAux := path
														copy(PathChars[:], PathAux)
														copy(BitacoraAux.Path[:], PathChars[:])
														//Setemos el propietario a la bitacora
														var ProperChars [16]byte
														copy(ProperChars[:], idSesion)
														copy(BitacoraAux.Proper[:], ProperChars[:])
														//Setemos el grupo a la bitacora
														var GrupoChars [16]byte
														copy(GrupoChars[:], idGrupo)
														copy(BitacoraAux.Grupo[:], GrupoChars[:])
														//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
														var OperacionChars [16]byte
														OperacionAux := "Edit"
														copy(OperacionChars[:], OperacionAux)
														copy(BitacoraAux.Operacion[:], OperacionChars[:])
														//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
														BitacoraAux.Tipo = 0
														//Setemos el contenido
														ContenidoEdit := cont
														var ContenidoChars [300]byte
														copy(ContenidoChars[:], ContenidoEdit)
														copy(BitacoraAux.Contenido[:], ContenidoChars[:])
														BitacoraAux.Size = int32(Fsize)
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
														color.Printf("@{w}El archivo @{w}%v @{w}fue editado con éxito\n", carpetas[len(carpetas)-1])

													} else {
														color.Println("@{r}El size debe ser mayor que cero.")
													}

												} else {
													PathCorrecto = false
													color.Printf("@{r} El usuario @{w}%v @{r}no tiene permisos de lectura y escritura sobre el archivo @{w}%v.\n", idSesion, carpetas[len(carpetas)-1])
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
								color.Println("@{r} Error, una o más carpetas padre no existen.")

							}

							fmt.Println("")

							///////////////////////////////////////////////////////////////////////////////////

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

												if sesionRoot || LecturaYEscrituraPropietarioFile(&InodoAux) || LecturaYEscrituraGrupoFile(&InodoAux, id) || LecturaYEscrituraOtrosFile(&InodoAux) {

													if EsCorrecto, Fsize := SizeCorrecto(size); EsCorrecto {

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
																InodoData := leerBytes(fileMBR, SizeInodo)
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
														//EN ESTA PARTE CADENACONTENIDO YA TIENE EL CONTENIDO
														contLeido := string(Contenido[:ContenidoSize])

														if cont != "" {
															contLeido = cont
														}

														//Antes de guardar el nuevo contenido debemos setear a cero los apuntadores y las posiciones en el bitmap
														//LEEMOS EL PRIMER INODO QUE ES EL INODO DEL ARCHIVO USERS.TXT
														InodoPointer := DDAux.DDFiles[i].ApuntadorInodo
														fileMBR.Seek(int64(InodoPointer+1), 0)
														InodoAux = estructuras.Inodo{}
														InodoSize := int(unsafe.Sizeof(InodoAux))
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
																	InodoAux.ApuntadoresBloques[i] = 0
																}
															}

															//Ahora toca escribir el ultimo strcut Inodo creado en su posición correspondiente
															fileMBR.Seek(int64(InodoPointer+1), 0)
															inodop := &InodoAux
															var binario bytes.Buffer
															binary.Write(&binario, binary.BigEndian, inodop)
															escribirBytes(fileMBR, binario.Bytes())

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

														InodoPointer = DDAux.DDFiles[i].ApuntadorInodo
														fileMBR.Seek(int64(InodoPointer+1), 0)
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

														contenido := ""
														FileSize := 0

														//Ahora toca escribir el contenido en los bloques de datos
														//y enlazarlos con newInodo, cabe recalcar que se debe validar
														//la cantidad de bloques necesarios, para crear un nuevo inodo de ser necesario

														if size == "" && contLeido == "" { //Si no vienen nungun parámetro ni size ni cont, el tamaño de 0

															//Si el size es cero, el archivo no contiene datos, por lo tanto no se le asigna ningun bloque
															//y podemos proceder a guardarlo inmediatamente

															InodoAux.FileSize = 0
															InodoAux.NumeroBloques = 0
															InodoAux.ApuntadorIndirecto = 0

														} else {

															if Fsize == 0 && size != "" {

																//Si el size es cero, el archivo no contiene datos, por lo tanto no se le asigna ningun bloque
																//y podemos proceder a guardarlo inmediatamente

																InodoAux.FileSize = 0
																InodoAux.NumeroBloques = 0
																InodoAux.ApuntadorIndirecto = 0

															} else if Fsize == 0 && size == "" && contLeido == "" {

																//Si el size es cero, el archivo no contiene datos, por lo tanto no se le asigna ningun bloque
																//y podemos proceder a guardarlo inmediatamente

																InodoAux.FileSize = 0
																InodoAux.NumeroBloques = 0
																InodoAux.ApuntadorIndirecto = 0

															} else {

																if contLeido == "" { //No hay contenido (setear el abcdario)

																	contenido = getNTimesabc(Fsize)
																	FileSize = Fsize

																} else if size == "" && contLeido != "" {

																	contenido = contLeido
																	FileSize = len(contenido)

																} else {

																	if Fsize > len(contLeido) { //size es mayor que el tamaño de cont, completar cont con el abcdario hasta llegar a size

																		FileSize = Fsize
																		contenido = contLeido        //contenido se le concatena el contenido enviado como parametro, que tiene un tamaño menor a size
																		r := Fsize - len(contLeido)  //calculamos cuantos caraceres hay que agregarle a contenido para cumplir con el size
																		contenido += getNTimesabc(r) //a contenido le enviamos los caraćteres del abcdario necesarios para llegar al size

																	} else if Fsize == len(contLeido) { //si size es igual al tamaño de cont, llenamos los bloques con cont

																		FileSize = Fsize
																		contenido = contLeido

																	} else if Fsize < len(contLeido) { //si size es menor que el tamaño de cont, cortamos cont hasta el tamaño de size

																		FileSize = Fsize
																		contenido = contLeido[:Fsize]

																	}

																}

															}

														}

														//Dividimos la cantidad de caracteres que tiene contenido dentro de 25
														//si el resultado es un numero decimal lo aproximamos al entero más cercano a la derecha
														//esto para saber cuando bloques de datos vamos a usar
														//Ejemplo: Si tenemos 52 caracteres, al dividirlo dentro de 25 obtenemos 2.08
														//ese 0.08 nos obliga a tomar un 3er bloque, entonces la funcion Roundf, aproximaria 2.08 a 3
														CadenaContenido := contenido
														FileSize = len(CadenaContenido)
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
																fileMBR.Seek(int64(InodoPointer+1), 0)
																inodop := &InodoAux
																var binario bytes.Buffer
																binary.Write(&binario, binary.BigEndian, inodop)
																escribirBytes(fileMBR, binario.Bytes())

																//PermisosAux
																valorU := InodoAux.PermisoU
																valorG := InodoAux.PermisoG
																valorO := InodoAux.PermisoO

																var ProperChars [16]byte
																copy(ProperChars[:], idSesion)
																//Setemos el grupo a la bitacora
																var GrupoChars [16]byte
																copy(GrupoChars[:], idGrupo)

																//Ahora creamos el nuevo inodo
																InodoAux = estructuras.Inodo{}
																InodoPointer = int32(Inodo2Pos)

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
																InodoAux.PermisoU = valorU
																InodoAux.PermisoG = valorG
																InodoAux.PermisoO = valorO
																copy(InodoAux.Proper[:], ProperChars[:])
																copy(InodoAux.Grupo[:], GrupoChars[:])

																//Seteamos SB
																SB1.FreeInodos = SB1.FreeInodos - 1

																indx = 0
															}
														}

														InodoAux.FileSize = int32(FileSize)
														InodoAux.NumeroBloques = CantidadBloques

														//Ahora toca escribir el ultimo strcut Inodo creado en su posición correspondiente
														fileMBR.Seek(int64(InodoPointer+1), 0)
														inodop := &InodoAux
														var binario bytes.Buffer
														binary.Write(&binario, binary.BigEndian, inodop)
														escribirBytes(fileMBR, binario.Bytes())

														//Crear bitacora MKUSR
														//Creamos la bitacora para la creación de la carpeta
														BitacoraAux := estructuras.Bitacora{}
														//Seteamos el path, en este caso la primera carpeta tiene "/" como path
														var PathChars [300]byte
														PathAux := path
														copy(PathChars[:], PathAux)
														copy(BitacoraAux.Path[:], PathChars[:])
														//Setemos el propietario a la bitacora
														var ProperChars [16]byte
														copy(ProperChars[:], idSesion)
														copy(BitacoraAux.Proper[:], ProperChars[:])
														//Setemos el grupo a la bitacora
														var GrupoChars [16]byte
														copy(GrupoChars[:], idGrupo)
														copy(BitacoraAux.Grupo[:], GrupoChars[:])
														//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
														var OperacionChars [16]byte
														OperacionAux := "Edit"
														copy(OperacionChars[:], OperacionAux)
														copy(BitacoraAux.Operacion[:], OperacionChars[:])
														//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
														BitacoraAux.Tipo = 0
														//Setemos el contenido
														ContenidoEdit := cont
														var ContenidoChars [300]byte
														copy(ContenidoChars[:], ContenidoEdit)
														copy(BitacoraAux.Contenido[:], ContenidoChars[:])
														BitacoraAux.Size = int32(Fsize)
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
														color.Printf("@{w}El archivo @{w}%v @{w}fue editado con éxito\n", carpetas[len(carpetas)-1])

													} else {
														color.Println("@{r}El size debe ser mayor que cero.")
													}

												} else {
													PathCorrecto = false
													color.Printf("@{r} El usuario @{w}%v @{r}no tiene permisos de lectura y escritura sobre el archivo @{w}%v.\n", idSesion, carpetas[len(carpetas)-1])
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
								color.Println("@{r} Error, una o más carpetas padre no existen.")

							}

							fmt.Println("")

							///////////////////////////////////////////////////////////////////////////////////

						} else {
							color.Println("@{r} La partición indicada no ha sido formateada.")
						}

						fileMBR.Close()

					}

				} else {
					color.Printf("@{r}No hay ninguna partición montada con el id: %v\n", id)
				}

			} else {
				color.Println("@{r}Path incorrecto, debe iniciar con @{w}/")
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios para la funcion REN.")
		}

	} else {
		color.Println("@{r}Se necesita de una sesión activa para ejecutar la función EDIT.")
	}

}
