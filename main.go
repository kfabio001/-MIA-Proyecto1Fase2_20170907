package main

import (
	"Proyecto1/analizadores"
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/doun/terminal/color"
)

//Pausa fuc
func Pausa() {
	color.Print("@{c}Ejecución pausada. Presiona 'Enter' para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

//LimpiarPantalla fuction
func LimpiarPantalla() {

	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	color.Println("@{!w}////////////////Manejo Implementacion archivos///////////////")
	color.Println("@{w}Ingrese comando exec:")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()

}

//TrimSuffix elimina el sufijo "suffix" de la cadena "s" si es que lo contiene
func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

//TrimPrefix elimina el prefijo "prefix" de la cadena "s" si es que lo contiene
func TrimPrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		s = s[len(prefix):]
	}
	return s
}

func main() {

	continuar := true
	LimpiarPantalla()

	for continuar {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("")
		//color.Print("@{!y}>>")
		input, _ := reader.ReadString('\n')

		if runtime.GOOS == "windows" {
			input = strings.TrimRight(input, "\r\n")
		} else {
			input = strings.TrimRight(input, "\n")
		}

		if strings.HasSuffix(input, "\\*") {
			pedir := true
			linea := ""
			for pedir {
				linea, _ = reader.ReadString('\n')
				if runtime.GOOS == "windows" {
					linea = strings.TrimRight(linea, "\r\n")
				} else {
					linea = strings.TrimRight(linea, "\n")
				}

				input += " "
				input += linea
				if !strings.HasSuffix(input, "\\*") {
					pedir = false
				}
			}
		}

		if strings.HasPrefix(strings.ToLower(input), "exec -path->") {

			path := input[12:]

			path = TrimSuffix(path, "\"")
			path = TrimPrefix(path, "\"")

			if path != "" {

				if strings.HasSuffix(strings.ToLower(path), ".arch") {

					if fileExists(path) {

						filebuffer, err := ioutil.ReadFile(path)
						if err != nil {
							color.Println("@{!r}Error al leer script.")
							fmt.Println(err)

						}

						analizadores.Lexico(string(filebuffer))

					} else {
						color.Println("@{!r}El script especificado no existe.")
					}

				} else {
					color.Println("@{!r}La ruta debe especificar un archivo con extension '.mia'.")
				}

			} else {
				color.Println("@{!r}La ruta no puede ser una cadena vacia.")
			}

		} else if strings.ToLower(input) == "pause" {
			Pausa()
		} else if strings.ToLower(input) == "exit" {
			continuar = false
		} else if strings.ToLower(input) == "clear" {
			LimpiarPantalla()
		} else {
			analizadores.Lexico(input)
		}
	}

	//comando := "mdisk create path->10M"
	/*filename := "hola.txt"
	filebuffer, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	analizadores.Lexico(string(filebuffer), &errorLexico)
	*/
	/*if !errorLexico {
		fmt.Println("Analisis lexico exitoso")
	} else {
		fmt.Println("Analisis lexico no exitoso")
	}*/

}
