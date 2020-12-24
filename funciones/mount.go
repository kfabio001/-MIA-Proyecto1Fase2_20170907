package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"github.com/doun/terminal/color"
)

var (
	//Discos donde se almacenaran los discos que tienen al menos una partición
	Discos []*estructuras.MD
	//PMList lista de todas las particionesv
	PMList []*estructuras.PM
)

const abc = "abcdefghijklmnopqrstuvwxyz"

//EjecutarMount function
func EjecutarMount(path string, name string) {

	if path != "" && name != "" {
		if fileExists(path) {

			existe, _ := ExisteParticion(path, name)
			existel, _ := ExisteParticionLogica(path, name)

			if existe || existel {

				if !EsExtendida(path, name) {

					if DiscoRegistrado, i := DiscoYaRegistrado(path); DiscoRegistrado {

						if ParticionRegistrada := ParticionYaRegistrada(path, name); !ParticionRegistrada {
							Discos[i].MDcount++
							id := "vd"
							id += getABC(i + 1)
							num := fmt.Sprint(Discos[i].MDcount)
							id += num
							Discos[i].Particiones = append(Discos[i].Particiones, id)

							newPM := new(estructuras.PM)
							newPM.PMid = id
							newPM.PMname = name
							newPM.PMpath = path
							PMList = append(PMList, newPM)
							CambiarStatusM(path, name, existe)
							color.Printf("@{g}Particion montada correctamente, el id asignado es: @{g}%v\n\n", id)
						} else {
							color.Println("@{r}Esta particion ya ha sido montada.")
						}

					} else {

						newReg := new(estructuras.MD)
						newReg.MDcount = 1
						newReg.MDocupado = 1
						newReg.MDpath = path
						Discos = append(Discos, newReg)
						id := "vd"
						id += getABC(len(Discos))
						id += "1"

						Discos[len(Discos)-1].Particiones = append(Discos[len(Discos)-1].Particiones, id)

						newPM := new(estructuras.PM)
						newPM.PMid = id
						newPM.PMname = name
						newPM.PMpath = path
						PMList = append(PMList, newPM)

						CambiarStatusM(path, name, existe)
						color.Printf("@{g}Particion montada correctamente, el id asignado es: @{g}%v\n\n", id)
					}

				} else {
					color.Println("@{r}No se puede montar porque es una partición extendida.")
				}

			} else {
				color.Println("@{r}El disco especificado no tiene ninguna partición con ese nombre.")
			}

		} else {
			color.Println("@{r}El disco especificado no existe.")
		}
	} else if path == "" && name == "" {
		DisplayPMList()
	} else {
		color.Println("@{r}Faltan parámetros obligatorios en la función MOUNT")
	}

}

//DiscoYaRegistrado verifica si ese disco ya tiene alguna otra particion montada, para asignar nueva letra
func DiscoYaRegistrado(path string) (bool, int) {

	if len(Discos) > 0 {
		for i := 0; i < len(Discos); i++ {
			if Discos[i].MDpath == path {
				return true, i
			}
		}
	}
	return false, 0
}

//ParticionYaRegistrada verifica si la partición ya ha sido montada con aterioridad
func ParticionYaRegistrada(path string, name string) bool {

	if len(PMList) > 0 {
		for i := 0; i < len(PMList); i++ {
			if PMList[i].PMpath == path && PMList[i].PMname == name {
				return true
			}
		}
	}
	return false
}

//IDYaRegistrado verifica si un id ya ha sido asignado a una particion ya montada
func IDYaRegistrado(id string) bool {

	if len(PMList) > 0 {
		for i := 0; i < len(PMList); i++ {
			if PMList[i].PMid == id {
				return true
			}
		}
	}
	return false
}

//GetDatosPart devuelve el name y el path
func GetDatosPart(id string) (string, string) {

	if len(PMList) > 0 {
		for i := 0; i < len(PMList); i++ {
			if PMList[i].PMid == id {
				return PMList[i].PMname, PMList[i].PMpath
			}
		}
	}
	return "", ""

}

func getABC(i int) string {
	return abc[i-1 : i]
}

//DisplayPMList funcion
func DisplayPMList() {

	if len(PMList) > 0 {
		fmt.Println("")
		color.Println("@{w}------ LISTA DE PARTICIONES MONTADAS ------")
		fmt.Println("")
		for _, pm := range PMList {
			//fmt.Printf("id->%v -path->%v -name->%v\n", pm.PMid, pm.PMpath, pm.PMname)
			color.Printf("@{b}id-> @{y}%v @{b}-path-> @{y}%v @{b}-name-> @{y}%v\n", pm.PMid, pm.PMpath, pm.PMname)
		}
		fmt.Println("")
	} else {
		color.Println("@{r}No hay ninguna partición montada hasta el momento.")
	}

}

//CambiarStatusM reescribe el atributos Status en el MBR o EBR de la particion
func CambiarStatusM(path string, name string, existe bool) {

	if existe {

		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
			file.Close()
		}

		Disco1 := estructuras.MBR{}
		DiskSize := int(unsafe.Sizeof(Disco1))
		file.Seek(0, 0)
		DiskData := leerBytes(file, DiskSize)
		buffer := bytes.NewBuffer(DiskData)
		err = binary.Read(buffer, binary.BigEndian, &Disco1)
		if err != nil {
			file.Close()
			panic(err)
		}
		for i := 0; i < 4; i++ {
			var chars [16]byte
			copy(chars[:], name)
			if string(Disco1.Mpartitions[i].Pname[:]) == string(chars[:]) {
				Disco1.Mpartitions[i].Pstatus = 'A'
			}
		}

		file.Seek(0, 0)
		m1 := &Disco1
		var binario bytes.Buffer
		binary.Write(&binario, binary.BigEndian, m1)
		escribirBytes(file, binario.Bytes())
		file.Close()

	}
}

//CambiarStatusU reescribe el atributos Status en el MBR o EBR de la particion
func CambiarStatusU(path string, name string) {

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	Disco1 := estructuras.MBR{}
	DiskSize := int(unsafe.Sizeof(Disco1))
	file.Seek(0, 0)
	DiskData := leerBytes(file, DiskSize)
	buffer := bytes.NewBuffer(DiskData)
	err = binary.Read(buffer, binary.BigEndian, &Disco1)
	if err != nil {
		file.Close()
		panic(err)
	}
	for i := 0; i < 4; i++ {
		var chars [16]byte
		copy(chars[:], name)
		if string(Disco1.Mpartitions[i].Pname[:]) == string(chars[:]) {
			Disco1.Mpartitions[i].Pstatus = 'D'
		}
	}

	file.Seek(0, 0)
	m1 := &Disco1
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, m1)
	escribirBytes(file, binario.Bytes())
	file.Close()

}

//GetID devuelve el id
func GetID(path string, name string) string {

	if len(PMList) > 0 {
		for i := 0; i < len(PMList); i++ {
			if PMList[i].PMpath == path && PMList[i].PMname == name {
				return PMList[i].PMid
			}
		}
	}
	return ""
}
