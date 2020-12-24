package funciones

import (
	"Proyecto1/estructuras"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarFDisk function
func EjecutarFDisk(size string, unit string, path string, tipo string, fit string, delete string, name string, add string) {
	color.Println(path)
	valorSize := 0
	valorBytes := 0

	if delete == "" && add == "" { //Fdisk normal (crea particiones)

		if size != "" && path != "" && name != "" {

			if strings.HasSuffix(strings.ToLower(path), ".disk") {

				if fileExists(path) {
					existe, _ := ExisteParticion(path, name)
					existel, _ := ExisteParticionLogica(path, name)

					if !existe && !existel {

						if i, _ := strconv.Atoi(size); i > 0 {

							valorSize = i

							if strings.ToLower(unit) == "k" || unit == "" {
								valorBytes = 1024
							} else if strings.ToLower(unit) == "b" {
								valorBytes = 1
							} else if strings.ToLower(unit) == "m" {
								valorBytes = 1024 * 1024
							}

							valorReal := valorSize * valorBytes
							CrearParticion(valorReal, path, tipo, fit, name)

						} else {
							color.Println("@{r}El size debe ser mayor que cero.")
						}

					} else {
						color.Println("@{r}Ya existe una particion con este nombre.")
					}

				} else {
					color.Println("@{r}El disco especificado no existe.")
				}

			} else {
				color.Println("@{r}La ruta debe especificar un archivo con extension '.dsk'.")
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios en la función FDISK")
		}

	} else if delete == "" && add != "" { // Fdisk para agregar o quitar espacio de una particion

	} else if delete != "" && add == "" { // Fdisk para eliminar una particion

		if path != "" && name != "" {

			if strings.HasSuffix(strings.ToLower(path), ".disk") {

				if fileExists(path) {

					if ExisteP, IndiceP := ExisteParticion(path, name); ExisteP {

						color.Printf("@{w}¿Está segur@@ que desea borrar esta partición?[Y/n]")

						pedir := true
						linea := ""

						for pedir {
							reader := bufio.NewReader(os.Stdin)
							input, _ := reader.ReadString('\n')

							if runtime.GOOS == "windows" {
								input = strings.TrimRight(input, "\r\n")
							} else {
								input = strings.TrimRight(input, "\n")
							}

							if strings.ToLower(input) == "n" || strings.ToLower(input) == "y" {
								linea = input
								pedir = false
							}

						}

						if strings.ToLower(linea) == "y" {

							if ParticionYaRegistrada(path, name) {
								idAux := GetID(path, name)
								Desmontar(path, name, idAux)
							}

							if strings.ToLower(delete) == "fast" {
								EliminacionFast(path, IndiceP)

							} else if strings.ToLower(delete) == "full" {
								EliminacionFull(path, IndiceP)
							}
							color.Println("@{w}Particion eliminada exitosamente.")
						}
						//AQUI VERIFICAMOS SI EXISTE UNA LOGICA
					} else if ExisteL, _ := ExisteParticionLogica(path, name); ExisteL {

					} else {
						color.Println("@{r}La particion no existe.")
					}

				} else {
					color.Println("@{r}El disco especificado no existe.")
				}

			} else {
				color.Println("@{r}La ruta debe especificar un archivo con extension '.disk'.")
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios en la función FDISK")
		}

	} else {
		color.Println("@{r}Los parámetros '-delete' y '-add' no pueden venir en la misma instruccion.")
	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()

}

//CrearParticion fuction
//size es el valor que debe tener la particion
func CrearParticion(size int, path string, tipo string, fit string, name string) {

	if PuedoAgregarParticion(path) {

		if strings.ToLower(tipo) == "p" || tipo == "" { // Particion primaria

			HayEspacio, Start := EspacioDisponible(size, path)

			if HayEspacio {

				fmt.Printf("La particion iniciara en el byte: %d\n", Start)

				Pindice := IndiceParticion(path)
				CrearPrimariaOExtendida(Pindice, Start, size, path, fit, name, tipo)
				color.Printf("@{w}La particion primaria @{w}%v @{w}fue creada con éxito\n", name)

			} else {
				color.Println("@{r}Operación fallida. No hay espacio disponible para nueva particion.")
			}

		} else if strings.ToLower(tipo) == "e" { // Particion extendida

			if !ExisteExtendida(path) {

				HayEspacio, Start := EspacioDisponible(size, path)

				if HayEspacio {

					fmt.Printf("La particion iniciara en el byte: %d\n", Start)

					Pindice := IndiceParticion(path)
					CrearPrimariaOExtendida(Pindice, Start, size, path, fit, name, tipo)
					color.Printf("@{w}La particion extendida @{w}%v @{w}fue creada con éxito\n", name)
				} else {
					color.Println("@{r}Operación fallida. No hay espacio disponible para nueva particion.")

				}

			} else {
				color.Println("@{r}El disco ha alcanzado el limite de particiones extendidas.")
			}

		} else if strings.ToLower(tipo) == "l" { // Particion logica
			if ExisteExtendida(path) {

				indiceExt, sizeExt := IndiceExtendida(path)

				if HayEspacio := CrearLogica(size, path, indiceExt, sizeExt, fit, name); HayEspacio {
					color.Printf("@{w}La particion lógica @{w}%v @{w}fue creada con éxito\n", name)
				} else {
					color.Println("@{r}Operación fallida. No hay espacio disponible para nueva particion.")
				}
			} else {
				color.Println("@{r}No se pudo crear la partición lógica porque el disco no tiene partición extendida.")
			}
		}

	} else if strings.ToLower(tipo) == "l" { // Particion logica
		if ExisteExtendida(path) {

			indiceExt, sizeExt := IndiceExtendida(path)

			if HayEspacio := CrearLogica(size, path, indiceExt, sizeExt, fit, name); HayEspacio {
				color.Printf("@{w}La particion lógica @{w}%v @{w}fue creada con éxito\n", name)

			} else {
				color.Println("@{r}Operación fallida. No hay espacio disponible para nueva particion.")
			}
		} else {
			color.Println("@{r}No se pudo crear la partición lógica porque el disco no tiene partición extendida.")
		}
	} else {
		color.Println("@{r}El disco ha alcanzado el limite de particiones.")
	}
}

//EspacioDisponible function, en caso de revolver TRUE, el valor entero es el byte de inicio para la nueva particion
func EspacioDisponible(size int, path string) (bool, int) {

	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	file.Seek(0, 0)

	for i := DiskSize; i <= int(Disco1.Msize)-size; i++ {
		vacio := true
		for x := 0; x < 4; x++ {

			if Disco1.Mpartitions[x].Psize > 0 {
				if i >= int(Disco1.Mpartitions[x].Pstart) && i <= int(Disco1.Mpartitions[x].Pstart)+int(Disco1.Mpartitions[x].Psize)-1 {
					vacio = false
				} else if i+size-1 >= int(Disco1.Mpartitions[x].Pstart) && i+size-1 <= int(Disco1.Mpartitions[x].Pstart)+int(Disco1.Mpartitions[x].Psize)-1 {
					vacio = false
				} else if i <= int(Disco1.Mpartitions[x].Pstart) && i+size-1 >= int(Disco1.Mpartitions[x].Pstart)+int(Disco1.Mpartitions[x].Psize)-1 {
					vacio = false
				} else if i == int(Disco1.Mpartitions[x].Pstart)+int(Disco1.Mpartitions[x].Psize)-1 {
					vacio = false
				} else if i+size-1 == int(Disco1.Mpartitions[x].Pstart) {
					vacio = false
				}
			}
		}

		if vacio {
			file.Close()
			return true, i
		}

	}

	file.Close()
	return false, 0
}

//PuedoAgregarParticion function
//siguiendo la teoria de particiones, verifica si hay 3 o menos particiones en el disco
func PuedoAgregarParticion(path string) bool {
	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {
		if Disco1.Mpartitions[i].Psize == 0 {
			file.Close()
			return true
		}
	}

	file.Close()
	return false
}

//ExisteExtendida function
//verifica si ya existe o no una particion extendida
func ExisteExtendida(path string) bool {
	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {
		if Disco1.Mpartitions[i].Ptype == 'e' || Disco1.Mpartitions[i].Ptype == 'E' {
			file.Close()
			return true
		}
	}

	file.Close()
	return false
}

//IndiceParticion function, busca una posicion para almacenar la info de la nueva particion
// en el arreglo de structs (particiones) del MBR.
func IndiceParticion(path string) int {
	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {
		if Disco1.Mpartitions[i].Psize == 0 {
			file.Close()
			return i
		}
	}
	file.Close()
	return 0

}

//CrearPrimariaOExtendida function
//indiceMBR es la posicion en el arreglo de structs del MBR, start es el parametro Pstart de c/ particion
func CrearPrimariaOExtendida(indiceMBR int, start int, size int, path string, fit string, name string, tipo string) {

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	//Disco1 sera el apuntador al struct MBR temporal
	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	//Seteando los atributos de la nueva particion en un struct del arreglo del MBR
	Disco1.Mpartitions[indiceMBR].Pstart = uint32(start)
	Disco1.Mpartitions[indiceMBR].Psize = uint32(size)
	var chars [16]byte
	copy(chars[:], name)
	copy(Disco1.Mpartitions[indiceMBR].Pname[:], chars[:])
	Disco1.Mpartitions[indiceMBR].Pstatus = 'D' // D = desactivada , A = activada

	if strings.ToLower(tipo) == "p" || tipo == "" {
		Disco1.Mpartitions[indiceMBR].Ptype = 'P'
	} else if strings.ToLower(tipo) == "e" {
		Disco1.Mpartitions[indiceMBR].Ptype = 'E'
	}

	if strings.ToLower(fit) == "wf" || fit == "" {
		Disco1.Mpartitions[indiceMBR].Pfit = 'W'
	} else if strings.ToLower(fit) == "bf" {
		Disco1.Mpartitions[indiceMBR].Pfit = 'B'
	} else if strings.ToLower(fit) == "ff" {
		Disco1.Mpartitions[indiceMBR].Pfit = 'F'
	}

	//Re-escribiendo MBR en el archivo binario (disco)
	file.Seek(0, 0)
	m1 := &Disco1
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, m1)
	escribirBytes(file, binario.Bytes())

	if strings.ToLower(tipo) == "e" {
		//CREAR Y ALMACENAR EBR
		e := estructuras.EBR{}
		e.Enext = -1
		e.Eprev = -1
		file.Seek(int64(start)+1, 0)
		ebr1 := &e
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, ebr1)
		escribirBytes(file, binario1.Bytes())
	}

	file.Close()
}

//CrearLogica function, en caso de revolver TRUE, el valor entero es el byte de inicio para la nueva particion
func CrearLogica(size int, path string, extStart int, extSize int, fit string, name string) bool {

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	//EBRaux sera el apuntador al struct EBR temporal
	EBRaux := estructuras.EBR{}
	EBRSize := int(unsafe.Sizeof(EBRaux))
	file.Seek(int64(extStart)+1, 0)
	EBRData := leerBytes(file, EBRSize)
	buffer := bytes.NewBuffer(EBRData)
	err = binary.Read(buffer, binary.BigEndian, &EBRaux)
	if err != nil {
		file.Close()
		panic(err)
	}

	Continuar := true

	for Continuar {

		if EBRaux.Enext == -1 {

			if EBRaux.Eprev == -1 { //primer EBR de la extendida, solo el primer EBR tendrá ValorPrev = -1

				if EBRaux.Esize == 0 { // esto significa que el primer EBr de la extendida no apunta a ninguna espacio (0 logicas)

					if int(extSize)-EBRSize >= size {
						EBRaux.Estart = uint32(extStart + EBRSize)
						EBRaux.Esize = uint32(size)
						EBRaux.Estatus = 'D'

						var chars [16]byte
						copy(chars[:], name)
						copy(EBRaux.Ename[:], chars[:])

						if strings.ToLower(fit) == "wf" || fit == "" {
							EBRaux.Efit = 'W'
						} else if strings.ToLower(fit) == "bf" {
							EBRaux.Efit = 'B'
						} else if strings.ToLower(fit) == "ff" {
							EBRaux.Efit = 'F'
						}

						file.Seek(int64(extStart)+1, 0)
						ebr1 := &EBRaux
						var binario1 bytes.Buffer
						binary.Write(&binario1, binary.BigEndian, ebr1)
						escribirBytes(file, binario1.Bytes())
						file.Close()
						return true
					}

				} else {

					if int(extSize)-int(EBRSize+int(EBRaux.Esize)) >= (EBRSize + size) {
						newEBR := estructuras.EBR{}
						newEBR.Enext = -1
						newEBR.Eprev = int32(extStart)
						newEBR.Estart = uint32(extStart + EBRSize + int(EBRaux.Esize) + EBRSize)
						newEBR.Esize = uint32(size)
						newEBR.Estatus = 'D'

						var chars [16]byte
						copy(chars[:], name)
						copy(newEBR.Ename[:], chars[:])

						if strings.ToLower(fit) == "wf" || fit == "" {
							newEBR.Efit = 'W'
						} else if strings.ToLower(fit) == "bf" {
							newEBR.Efit = 'B'
						} else if strings.ToLower(fit) == "ff" {
							newEBR.Efit = 'F'
						}

						file.Seek(int64(extStart+EBRSize+int(EBRaux.Esize)+1), 0)
						ebr1 := &newEBR
						var binario1 bytes.Buffer
						binary.Write(&binario1, binary.BigEndian, ebr1)
						escribirBytes(file, binario1.Bytes())

						EBRaux.Enext = int32(extStart + EBRSize + int(EBRaux.Esize))
						file.Seek(int64(extStart)+1, 0)
						ebr1 = &EBRaux
						var binario2 bytes.Buffer
						binary.Write(&binario2, binary.BigEndian, ebr1)
						escribirBytes(file, binario2.Bytes())
						file.Close()
						return true

					}

				}
			} else {

				if (extStart+extSize)-int(EBRaux.Estart+EBRaux.Esize) >= (EBRSize + size) {

					newEBR := estructuras.EBR{}
					newEBR.Enext = -1
					newEBR.Eprev = int32(int32(EBRaux.Estart) - int32(EBRSize))
					newEBR.Estart = uint32(EBRaux.Estart + EBRaux.Esize + uint32(EBRSize))
					newEBR.Esize = uint32(size)
					newEBR.Estatus = 'D'

					var chars [16]byte
					copy(chars[:], name)
					copy(newEBR.Ename[:], chars[:])

					if strings.ToLower(fit) == "wf" || fit == "" {
						newEBR.Efit = 'W'
					} else if strings.ToLower(fit) == "bf" {
						newEBR.Efit = 'B'
					} else if strings.ToLower(fit) == "ff" {
						newEBR.Efit = 'F'
					}

					file.Seek(int64(EBRaux.Estart+EBRaux.Esize+1), 0)
					ebr1 := &newEBR
					var binario1 bytes.Buffer
					binary.Write(&binario1, binary.BigEndian, ebr1)
					escribirBytes(file, binario1.Bytes())

					EBRaux.Enext = int32(EBRaux.Estart + EBRaux.Esize)
					file.Seek(int64(int32(EBRaux.Estart)-int32(EBRSize)+1), 0)
					ebr1 = &EBRaux
					var binario2 bytes.Buffer
					binary.Write(&binario2, binary.BigEndian, ebr1)
					escribirBytes(file, binario2.Bytes())
					file.Close()
					return true
				}

			}

		} else {

			if EBRaux.Eprev == -1 && EBRaux.Esize == 0 {

				if extStart+EBRSize+size < int(EBRaux.Enext) {
					EBRaux.Estart = uint32(extStart + EBRSize)
					EBRaux.Esize = uint32(size)
					EBRaux.Estatus = 'D'

					var chars [16]byte
					copy(chars[:], name)
					copy(EBRaux.Ename[:], chars[:])

					if strings.ToLower(fit) == "wf" || fit == "" {
						EBRaux.Efit = 'W'
					} else if strings.ToLower(fit) == "bf" {
						EBRaux.Efit = 'B'
					} else if strings.ToLower(fit) == "ff" {
						EBRaux.Efit = 'F'
					}

					file.Seek(int64(extStart)+1, 0)
					ebr1 := &EBRaux
					var binario1 bytes.Buffer
					binary.Write(&binario1, binary.BigEndian, ebr1)
					escribirBytes(file, binario1.Bytes())
					file.Close()
					return true

				}

			} else {

				if EBRaux.Enext-int32(EBRaux.Estart+EBRaux.Esize) >= int32(EBRSize+size) {
					NextTemporal := EBRaux.Enext
					//CREANDO EL NUEVO EBR
					newEBR := estructuras.EBR{}
					newEBR.Enext = NextTemporal
					newEBR.Eprev = int32(int32(EBRaux.Estart) - int32(EBRSize))
					newEBR.Estart = uint32(EBRaux.Estart + EBRaux.Esize + uint32(EBRSize))
					newEBR.Esize = uint32(size)
					newEBR.Estatus = 'D'

					var chars [16]byte
					copy(chars[:], name)
					copy(newEBR.Ename[:], chars[:])

					if strings.ToLower(fit) == "wf" || fit == "" {
						newEBR.Efit = 'W'
					} else if strings.ToLower(fit) == "bf" {
						newEBR.Efit = 'B'
					} else if strings.ToLower(fit) == "ff" {
						newEBR.Efit = 'F'
					}
					//GUARDANDO EL NUEVO EBR
					file.Seek(int64(EBRaux.Estart+EBRaux.Esize+1), 0)
					ebr1 := &newEBR
					var binario1 bytes.Buffer
					binary.Write(&binario1, binary.BigEndian, ebr1)
					escribirBytes(file, binario1.Bytes())
					//REESCRIBIENDO EL EBR ACTUAL
					EBRaux.Enext = int32(EBRaux.Estart + EBRaux.Esize)
					file.Seek(int64(int32(EBRaux.Estart)-int32(EBRSize)+1), 0)
					ebr1 = &EBRaux
					var binario2 bytes.Buffer
					binary.Write(&binario2, binary.BigEndian, ebr1)
					escribirBytes(file, binario2.Bytes())
					//LEYENDO Y REESCRIBIENDO EL EBR SIGUIENTE
					EBRsig := estructuras.EBR{}

					EBRSize = int(unsafe.Sizeof(EBRsig))
					file.Seek(int64(NextTemporal)+1, 0)
					EBRData2 := leerBytes(file, EBRSize)
					buffer2 := bytes.NewBuffer(EBRData2)

					err = binary.Read(buffer2, binary.BigEndian, &EBRsig)
					if err != nil {
						file.Close()
						panic(err)
					}

					EBRsig.Eprev = int32(newEBR.Estart - uint32(EBRSize))
					file.Seek(int64(NextTemporal)+1, 0)
					ebr1 = &EBRsig
					var binario3 bytes.Buffer
					binary.Write(&binario3, binary.BigEndian, ebr1)
					escribirBytes(file, binario3.Bytes())
					file.Close()
					return true
				}
			}

		}

		if EBRaux.Enext != -1 {
			//Si hay otro EBR a la derecha lo leemos y volvemos al inicio del FOR
			file.Seek(int64(EBRaux.Enext)+1, 0)
			EBRData := leerBytes(file, EBRSize)
			buffer := bytes.NewBuffer(EBRData)
			err = binary.Read(buffer, binary.BigEndian, &EBRaux)
			if err != nil {
				file.Close()
				panic(err)
			}
		} else {
			//Si no cancelamos, por lo tanto no hay espacio y retornara FALSE
			Continuar = false
		}

	}

	file.Close()
	return false
}

//IndiceExtendida function
//el primer valor int que retorna es el byte donde inicia la particion extendida
//este metodo solamente se llama cuando se ha verificado que existe una extendida
//por lo tanto nunca retorna 0
//el segundo valor int es el size de la extendida
func IndiceExtendida(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {
		if Disco1.Mpartitions[i].Ptype == 'e' || Disco1.Mpartitions[i].Ptype == 'E' {
			file.Close()
			return int(Disco1.Mpartitions[i].Pstart), int(Disco1.Mpartitions[i].Psize)
		}
	}

	file.Close()
	return 0, 0
}

//ExisteParticion function
//verifica si ya existe o no una particion extendida
//devuelve el indice en el arreglo de particiones del mbr
func ExisteParticion(path string, name string) (bool, int) {
	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {

		if Disco1.Mpartitions[i].Psize > 0 {
			var chars [16]byte
			copy(chars[:], name)
			if string(Disco1.Mpartitions[i].Pname[:]) == string(chars[:]) {
				file.Close()
				return true, i
			}
		}
	}

	file.Close()
	return false, 0
}

// EliminacionFast function
func EliminacionFast(path string, indiceMBR int) {

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}

	Disco1.Mpartitions[indiceMBR].Psize = 0
	Disco1.Mpartitions[indiceMBR].Ptype = 0
	//Re-escribiendo MBR en el archivo binario (disco)
	file.Seek(0, 0)
	m1 := &Disco1
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, m1)
	escribirBytes(file, binario.Bytes())

	file.Close()
}

// EliminacionFull function
func EliminacionFull(path string, indiceMBR int) {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	//reescribrimos un arreglo de bytes vacios en el lugar donde iba la particion borrada
	file.Seek(int64(Disco1.Mpartitions[indiceMBR].Pstart)+1, 0)
	data := make([]byte, int(Disco1.Mpartitions[indiceMBR].Psize))
	escribirBytes(file, data)
	//reseteamos los atributos del struct en el arreglo del MBR
	Disco1.Mpartitions[indiceMBR].Psize = 0
	Disco1.Mpartitions[indiceMBR].Ptype = 0
	Disco1.Mpartitions[indiceMBR].Pfit = 0
	Disco1.Mpartitions[indiceMBR].Ptype = 0
	//Re-escribiendo MBR en el archivo binario (disco)
	file.Seek(0, 0)
	m1 := &Disco1
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, m1)
	escribirBytes(file, binario.Bytes())

	file.Close()
}

//ExisteParticionLogica function, devuelve True si la partición lógica existe y el byte donde inicia el EBR
func ExisteParticionLogica(path string, name string) (bool, int) {

	if ExisteExtendida(path) {

		indiceExt, _ := IndiceExtendida(path)

		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
			file.Close()
		}

		//EBRaux sera el apuntador al struct EBR temporal
		EBRaux := estructuras.EBR{}
		//Obtenemos el tamanio del ebr
		EBRSize := int(unsafe.Sizeof(EBRaux))
		file.Seek(int64(indiceExt)+1, 0)
		//Lee la cantidad de <size> bytes del archivo
		EBRData := leerBytes(file, EBRSize)
		//Convierte la data en un buffer,necesario para
		//decodificar binario
		buffer := bytes.NewBuffer(EBRData)

		//Decodificamos y guardamos en la variable EBRaux
		err = binary.Read(buffer, binary.BigEndian, &EBRaux)
		if err != nil {
			file.Close()
			panic(err)
		}

		Continuar := true

		for Continuar {

			var chars [16]byte
			copy(chars[:], name)
			if string(EBRaux.Ename[:]) == string(chars[:]) {
				file.Close()
				return true, int(int(EBRaux.Estart) - EBRSize)
			}

			if EBRaux.Enext != -1 {
				//Si hay otro EBR a la derecha lo leemos y volvemos al inicio del FOR
				file.Seek(int64(EBRaux.Enext)+1, 0)
				EBRData := leerBytes(file, EBRSize)
				buffer := bytes.NewBuffer(EBRData)
				err = binary.Read(buffer, binary.BigEndian, &EBRaux)
				if err != nil {
					file.Close()
					panic(err)
				}
			} else {
				//Si no cancelamos, por lo tanto no existe la particion logica.
				Continuar = false
			}

		}
		file.Close()
		return false, 0
	}
	return false, 0

}

//EsExtendida function
func EsExtendida(path string, name string) bool {
	file, err := os.Open(path)
	if err != nil { //validar que no sea nulo.
		panic(err)
	}

	Disco1 := estructuras.MBR{}
	//Obtenemos el tamanio del mbr
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	//Lee la cantidad de <size> bytes del archivo
	DiskData := leerBytes(file, DiskSize)
	//Convierte la data en un buffer,necesario para
	//decodificar binario
	buffer := bytes.NewBuffer(DiskData)

	//Decodificamos y guardamos en la variable Disco1
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {

		if Disco1.Mpartitions[i].Psize > 0 {
			var chars [16]byte
			copy(chars[:], name)
			if string(Disco1.Mpartitions[i].Pname[:]) == string(chars[:]) {
				if Disco1.Mpartitions[i].Ptype == 'e' || Disco1.Mpartitions[i].Ptype == 'E' {
					file.Close()
					return true
				}
			}
		}
	}

	file.Close()
	return false
}

//GetStartAndSize devuelve el byte donde inicia la particion y su size
func GetStartAndSize(path string, indice int) (int, int) {

	if indice >= 0 && indice <= 3 {

		file, err := os.Open(path)
		if err != nil { //validar que no sea nulo.
			panic(err)
		}

		Disco1 := estructuras.MBR{}
		//Obtenemos el tamanio del mbr
		DiskSize := int(unsafe.Sizeof(Disco1))
		file.Seek(0, 0)
		//Lee la cantidad de <size> bytes del archivo
		DiskData := leerBytes(file, DiskSize)
		//Convierte la data en un buffer,necesario para
		//decodificar binario
		buffer := bytes.NewBuffer(DiskData)

		//Decodificamos y guardamos en la variable Disco1
		err = binary.Read(buffer, binary.BigEndian, &Disco1)
		if err != nil {
			file.Close()
			panic(err)
		}

		file.Close()
		return int(Disco1.Mpartitions[indice].Pstart), int(Disco1.Mpartitions[indice].Psize)
	}

	return 0, 0

}

//CantidadLogicas devuelve la cantidad de particiones lógicas que tiene una extendida
func CantidadLogicas(path string) int {

	if ExisteExtendida(path) {

		indiceExt, _ := IndiceExtendida(path)

		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
			file.Close()
		}

		//EBRaux sera el apuntador al struct EBR temporal
		EBRaux := estructuras.EBR{}
		//Obtenemos el tamanio del ebr
		EBRSize := int(unsafe.Sizeof(EBRaux))
		file.Seek(int64(indiceExt)+1, 0)
		//Lee la cantidad de <size> bytes del archivo
		EBRData := leerBytes(file, EBRSize)
		//Convierte la data en un buffer,necesario para
		//decodificar binario
		buffer := bytes.NewBuffer(EBRData)

		//Decodificamos y guardamos en la variable EBRaux
		err = binary.Read(buffer, binary.BigEndian, &EBRaux)
		if err != nil {
			file.Close()
			panic(err)
		}

		Cantidad := 0

		Continuar := true

		for Continuar {

			if EBRaux.Esize > 0 {
				Cantidad++
			}

			if EBRaux.Enext != -1 {
				//Si hay otro EBR a la derecha lo leemos y volvemos al inicio del FOR
				file.Seek(int64(EBRaux.Enext)+1, 0)
				EBRData := leerBytes(file, EBRSize)
				buffer := bytes.NewBuffer(EBRData)
				err = binary.Read(buffer, binary.BigEndian, &EBRaux)
				if err != nil {
					file.Close()
					panic(err)
				}
			} else {
				//Si no cancelamos, por lo tanto no existe la particion logica.
				Continuar = false
			}

		}
		file.Close()
		return Cantidad
	}
	return 0

}
