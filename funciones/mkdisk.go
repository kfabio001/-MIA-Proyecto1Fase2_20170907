package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/doun/terminal/color"
)

var juntar2 string

//EjecutarMkDisk function
func EjecutarMkDisk(size string, path string, name string, unit string) {
	var pathN []string
	var juntar string
	color.Println(path + " ! " + name)
	pathN = strings.Split(path, "/")
	for _, c := range pathN {
		//pathN = strings.Split(c, ".")
		//if len(pathN) > 1 {
		juntar2 = juntar
		juntar += c + "/"
		name = c
		//}
	}
	path = juntar2

	valorSize := 0
	valorBytes := 0

	if size != "" && path != "" && name != "" {

		if strings.HasSuffix(strings.ToLower(name), ".disk") {

			if i, _ := strconv.Atoi(size); i > 0 {

				valorSize = i
				if err := ensureDir(path); err == nil {

					if strings.ToLower(unit) == "m" || unit == "" || strings.ToLower(unit) == "k" {

						fullName := path + name

						if !fileExists(fullName) {
							if strings.ToLower(unit) == "m" || unit == "" {
								valorBytes = 1024 * 1024
							} else {
								valorBytes = 1024
							}

							valorReal := valorSize * valorBytes

							file, err := os.Create(fullName) //Crea un nuevo archivo
							if err != nil {
								panic(err)
							}

							// Change permissions Linux.
							err = os.Chmod(fullName, 0666)
							if err != nil {
								log.Println(err)
							}

							data := make([]byte, valorReal) //-size=2 -unit=K
							file.Write(data)                //Escribir datos como un arreglo de bytes

							// Convirtiendo string "valorreal" a uint32
							mUsize := uint32(valorReal)
							//Creando nuevo mbr
							s := estructuras.MBR{}
							//Asignando valor Msize (uint32)
							s.Msize = mUsize
							//Obteniendo fecha y hora actual, guardando como cadena, y asignando como Mdate
							var chars [20]byte
							t := time.Now()
							cadena := t.Format("2006-01-02 15:04:05")
							copy(chars[:], cadena)
							copy(s.Mdate[:], chars[:])
							//Generando valor random y asignando como Msignature
							s.Msignature = rand.Uint32()
							//Escribiendo MBR en el archivo binario (disco)
							file.Seek(0, 0)
							m1 := &s
							var binario bytes.Buffer
							binary.Write(&binario, binary.BigEndian, m1)
							escribirBytes(file, binario.Bytes())
							file.Close()
							color.Printf("@{w}El disco @{w}%v @{w}ha sido creado.\n", name)
						} else {
							color.Println("@{r}Este disco ya existe.")
						}

					} else {
						color.Println("@{r}Parámetro 'unit' inválido.")
					}

				} else {
					fmt.Println("Directorio inválido")
					fmt.Println(err)
				}

			} else {
				color.Println("@{r}El size debe ser mayor que cero.")
			}

		} else {
			color.Println("@{r}El nombre debe contener la extension '.disk'.")
		}

	} else {
		color.Println("@{r}Faltan parámetros obligatorios en la COMANDO MKDISK")
	}
}

func ensureDir(dirName string) error {

	err := os.MkdirAll(dirName, 0777)

	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func escribirBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)

	if err != nil {
		file.Close()
		panic(err)
	}
}

func leerBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number) //array de bytes

	_, err := file.Read(bytes) // Leido -> bytes
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}
