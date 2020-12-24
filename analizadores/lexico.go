package analizadores

import (
	"Proyecto1/estructuras"
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/doun/terminal/color"
)

var (
	tokens                           = []*estructuras.Token{}
	repetir, anular, errorLex bool   = false, false, false
	estado                    int    = 0
	lexemaact, lexemaant      string = "", ""
)

//Pausa fuc
func Pausa() {
	fmt.Println("")
	color.Print("@{c}Ejecución pausada. Presiona 'Enter' para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

var mkdi bool = false
var fitt bool = false
var fitt2 bool = false
var idee bool = false

// Lexico -> funcion para analizar cadena de entrada
// como parametro recibe una cadena ya sea ingresa en consola
// o el texto plano contenido en archivo leido
func Lexico(entrada string) {
	tokens = nil
	errorLex = false
	entrada += " \n"
	split := strings.Split(entrada, "")
	for _, c := range split {
		anular = false
		for ok := true; ok; ok = repetir {
			repetir = false
			switch estado {
			case 0: //posibles transiciones para el Estado 0 (Aceptacion de TK_ESPACIO)

				if isWhiteSpace(c) {
					continue
				} else if c == "\"" {
					lexemaact = ""

					estado = 1

				} else if isInt(c) {
					estado = 2
					lexemaact = c
				} else if c == "+" {
					estado = 3
					lexemaact = c
				} else if c == "*" {
					lexemaact = c
					newToken := estructuras.NewToken("TK_ASTERISCO", lexemaact)
					tokens = append(tokens, newToken)
					lexemaact = ""
					estado = 0
				} else if c == "-" {
					estado = 4
					lexemaact = c
				} else if isLetter(c) {
					estado = 9
					lexemaact = c
				} else if c == "/" {
					estado = 6
					lexemaact = c
				} else if c == "\\" {
					estado = 7
				} else if c == "#" {
					estado = 8
					lexemaact = c
				} else {
					errorLex = true
					lexemaact = ""
				}

			case 1: //posibles transiciones para el estado 1 (Cadenas)

				if c != "\"" {
					estado = 1
					lexemaact += c
				} else if c == "\"" {

					estado = 0
					//color.Println(lexemaact) rm
					if (strings.HasSuffix(strings.ToLower(lexemaact), ".disk") || strings.HasSuffix(strings.ToLower(lexemaact), ".arch") || strings.HasSuffix(strings.ToLower(lexemaact), ".pdf") || strings.HasSuffix(strings.ToLower(lexemaact), ".jpg") || strings.HasSuffix(strings.ToLower(lexemaact), ".txt") || strings.HasSuffix(strings.ToLower(lexemaact), ".png") || strings.HasSuffix(strings.ToLower(lexemaact), ".sh")) && mkdi == false {
						newToken := estructuras.NewToken("TK_FILE", lexemaact)

						tokens = append(tokens, newToken)
					} else {
						if idee == true {
							//color.Println(lexemaact)
							newToken := estructuras.NewToken("TK_ID", lexemaact)
							tokens = append(tokens, newToken)
							idee = false
						} else {
							newToken := estructuras.NewToken("TK_DIR", lexemaact)
							tokens = append(tokens, newToken)
						}
					}
					lexemaact = ""
				}

			case 2: //posibles transiciones para el estado 2 (numeros enteros)
				if isInt(c) {
					estado = 2
					lexemaact += c
				} else {
					newToken := estructuras.NewToken("TK_NUM", lexemaact)
					tokens = append(tokens, newToken)
					lexemaact = ""
					estado = 0
					repetir = true
				}

			case 3: //posibles transiciones para el estado 3 (numeros enteros con positivos)

				if isInt(c) {
					estado = 2
					lexemaact += c
				} else {
					errorLex = true
					lexemaact = ""
					estado = 0
					repetir = true
				}

			case 4: //posibles transiciones para el estado 4 (numeros negativos, palabras reservadas o flecha asignacion)

				if isInt(c) {
					estado = 2
					lexemaact += c
				} else if c == ">" {
					if fitt2 == true {
						lexemaact = ""
						//estado=0
					} else {
						lexemaact += c
						newToken := estructuras.NewToken("TK_ASIG", lexemaact)
						tokens = append(tokens, newToken)
						lexemaact = ""
					}
					estado = 0
				} else if isLetter(c) {
					estado = 5
					lexemaact += c
				} else {
					errorLex = true
					lexemaact = ""
					estado = 0
					repetir = true
				}

			case 5: //posibles transiciones para el estado 5 (palabras reservadas)

				if isLetter(c) {
					estado = 5
					lexemaact += c
				} else {

					switch strings.ToLower(lexemaact) {

					case "-path":
						newToken := estructuras.NewToken("TK_PATH", lexemaact)
						tokens = append(tokens, newToken)
					case "-size":
						newToken := estructuras.NewToken("TK_SIZE", lexemaact)
						tokens = append(tokens, newToken)
					case "-name":
						idee = true
						newToken := estructuras.NewToken("TK_NAME", lexemaact)
						//color.Println(lexemaact)
						tokens = append(tokens, newToken)
					case "-unit":
						newToken := estructuras.NewToken("TK_UNIT", lexemaact)
						tokens = append(tokens, newToken)
					case "fdisk":
						newToken := estructuras.NewToken("TK_FDISK", lexemaact)
						tokens = append(tokens, newToken)
					case "-type":
						newToken := estructuras.NewToken("TK_TYPE", lexemaact)
						tokens = append(tokens, newToken)
					case "-fit":
						if fitt == true {
							lexemaact = ""
							fitt2 = true
						} else {
							newToken := estructuras.NewToken("TK_FIT", lexemaact)
							tokens = append(tokens, newToken)
						}
					case "-delete":
						newToken := estructuras.NewToken("TK_DEL", lexemaact)
						tokens = append(tokens, newToken)
					case "-add":
						newToken := estructuras.NewToken("TK_ADD", lexemaact)
						tokens = append(tokens, newToken)
					case "-id":
						newToken := estructuras.NewToken("TK_PID", lexemaact)
						idee = true
						tokens = append(tokens, newToken)
					case "-nombre":
						newToken := estructuras.NewToken("TK_NOMBRE", lexemaact)
						tokens = append(tokens, newToken)
					case "-ruta":
						newToken := estructuras.NewToken("TK_RUTA", lexemaact)
						tokens = append(tokens, newToken)
					case "-usr":
						idee = true
						newToken := estructuras.NewToken("TK_USR", lexemaact)
						tokens = append(tokens, newToken)
					case "-pwd":
						newToken := estructuras.NewToken("TK_PWD", lexemaact)
						tokens = append(tokens, newToken)
					case "-grp":
						idee = true
						newToken := estructuras.NewToken("TK_GRP", lexemaact)
						tokens = append(tokens, newToken)
					case "-ugo":
						newToken := estructuras.NewToken("TK_UGO", lexemaact)
						tokens = append(tokens, newToken)
					case "-r":
						newToken := estructuras.NewToken("TK_R", lexemaact)
						tokens = append(tokens, newToken)
					case "-p":
						newToken := estructuras.NewToken("TK_P", lexemaact)
						tokens = append(tokens, newToken)
					case "-cont":
						newToken := estructuras.NewToken("TK_CONT", lexemaact)
						tokens = append(tokens, newToken)
					case "-file":
						newToken := estructuras.NewToken("TK_PFILE", lexemaact)
						tokens = append(tokens, newToken)
					case "-rf":
						newToken := estructuras.NewToken("TK_RF", lexemaact)
						tokens = append(tokens, newToken)
					case "-dest":
						newToken := estructuras.NewToken("TK_DEST", lexemaact)
						tokens = append(tokens, newToken)
					case "-iddestiny":
						newToken := estructuras.NewToken("TK_IDDEST", lexemaact)
						tokens = append(tokens, newToken)
					default:
						errorLex = true
						lexemaact = ""
						estado = 0
						repetir = true

					}
					estado = 0
					if !anular {
						repetir = true
					} else if anular {
						repetir = false
						anular = false
					}
					lexemaact = ""
				}

			case 6: //posibles transiciones para el estado 6 (rutas y file names sin comillas)

				if !isWhiteSpace(c) {
					estado = 6
					lexemaact += c
				} else {

					if (strings.HasSuffix(strings.ToLower(lexemaact), ".disk") || strings.HasSuffix(strings.ToLower(lexemaact), ".arch") || strings.HasSuffix(strings.ToLower(lexemaact), ".pdf") || strings.HasSuffix(strings.ToLower(lexemaact), ".jpg") || strings.HasSuffix(strings.ToLower(lexemaact), ".txt") || strings.HasSuffix(strings.ToLower(lexemaact), ".png") || strings.HasSuffix(strings.ToLower(lexemaact), ".txt") || strings.HasSuffix(strings.ToLower(lexemaact), ".sh")) && mkdi == false {
						newToken := estructuras.NewToken("TK_FILE", lexemaact)
						tokens = append(tokens, newToken)
					} else {
						//color.Println(lexemaact)
						newToken := estructuras.NewToken("TK_DIR", lexemaact)
						tokens = append(tokens, newToken)
					}
					lexemaact = ""
					estado = 0
					repetir = true
				}

			case 7: //posibles transiciones para el estado 7 (\*)
				if c != "*" {
					errorLex = true
					lexemaact = ""
					estado = 0
					repetir = true
				} else {
					lexemaact = ""
					estado = 0
				}
			case 8: //posibles transiciones para el estado 8 (comentarios) ff

				if c != "\n" {
					estado = 8
					lexemaact += c
				} else {
					newToken := estructuras.NewToken("TK_CMT", lexemaact)
					tokens = append(tokens, newToken)
					lexemaact = ""
					estado = 0
				}

			case 9: //palabras reservadas y ids

				if isLetter(c) || isInt(c) || c == "_" || c == "." {
					estado = 9
					lexemaact += c
				} else {

					switch strings.ToLower(lexemaact) {

					case "exec":
						newToken := estructuras.NewToken("TK_EXEC", lexemaact)
						tokens = append(tokens, newToken)
					case "pause":
						newToken := estructuras.NewToken("TK_PAUSE", lexemaact)
						tokens = append(tokens, newToken)
					case "mkdisk":
						newToken := estructuras.NewToken("TK_MKDISK", lexemaact)
						tokens = append(tokens, newToken)
						mkdi = true
						fitt = true
					case "rmdisk":
						newToken := estructuras.NewToken("TK_RMDISK", lexemaact)
						tokens = append(tokens, newToken)
					case "fdisk":
						fitt = false
						fitt2 = false
						mkdi = false
						newToken := estructuras.NewToken("TK_FDISK", lexemaact)
						tokens = append(tokens, newToken)
					case "mount":
						newToken := estructuras.NewToken("TK_MNT", lexemaact)
						tokens = append(tokens, newToken)
					case "unmount":
						newToken := estructuras.NewToken("TK_UMNT", lexemaact)
						tokens = append(tokens, newToken)
					case "b", "m", "k":
						newToken := estructuras.NewToken("TK_BYTES", lexemaact)
						tokens = append(tokens, newToken)
					case "p", "e", "l":
						newToken := estructuras.NewToken("TK_PEL", lexemaact)
						tokens = append(tokens, newToken)
					case "bf", "ff", "wf":
						if fitt == true {
							lexemaact = ""
						} else {

							newToken := estructuras.NewToken("TK_BFW", lexemaact)
							tokens = append(tokens, newToken)
						}
					case "fast", "full":
						newToken := estructuras.NewToken("TK_FF", lexemaact)
						tokens = append(tokens, newToken)
					case "mbr", "disk", "sb", "bm_arbdir", "bm_detdir", "bm_inode", "bm_block", "journaling", "block", "blocki", "inode", "tree", "ls":
						newToken := estructuras.NewToken("TK_TIPOREPORTE", lexemaact)
						tokens = append(tokens, newToken)
					//FASE 2
					case "rep":
						newToken := estructuras.NewToken("TK_REP", lexemaact)
						tokens = append(tokens, newToken)
					case "mkfs":
						newToken := estructuras.NewToken("TK_MKFS", lexemaact)
						tokens = append(tokens, newToken)
					case "login":
						newToken := estructuras.NewToken("TK_LOGIN", lexemaact)
						tokens = append(tokens, newToken)
					case "logout":
						newToken := estructuras.NewToken("TK_LOGOUT", lexemaact)
						tokens = append(tokens, newToken)
					case "mkgrp":
						newToken := estructuras.NewToken("TK_MKGRP", lexemaact)
						tokens = append(tokens, newToken)
					case "rmgrp":
						newToken := estructuras.NewToken("TK_RMGRP", lexemaact)
						tokens = append(tokens, newToken)
					case "mkusr":
						newToken := estructuras.NewToken("TK_MKUSR", lexemaact)
						tokens = append(tokens, newToken)
					case "rmusr":
						newToken := estructuras.NewToken("TK_RMUSR", lexemaact)
						tokens = append(tokens, newToken)
					case "chmod":
						newToken := estructuras.NewToken("TK_CHMOD", lexemaact)
						tokens = append(tokens, newToken)
					case "mkfile":
						newToken := estructuras.NewToken("TK_MKFILE", lexemaact)
						tokens = append(tokens, newToken)
					case "cat":
						newToken := estructuras.NewToken("TK_CAT", lexemaact)
						tokens = append(tokens, newToken)
					case "rem":
						newToken := estructuras.NewToken("TK_RM", lexemaact)
						tokens = append(tokens, newToken)
					case "edit":
						newToken := estructuras.NewToken("TK_EDIT", lexemaact)
						tokens = append(tokens, newToken)
					case "ren":
						newToken := estructuras.NewToken("TK_REN", lexemaact)
						tokens = append(tokens, newToken)
					case "mkdir":
						newToken := estructuras.NewToken("TK_MKDIR", lexemaact)
						tokens = append(tokens, newToken)
					case "cp":
						newToken := estructuras.NewToken("TK_CP", lexemaact)
						tokens = append(tokens, newToken)
					case "mv":
						newToken := estructuras.NewToken("TK_MV", lexemaact)
						tokens = append(tokens, newToken)
					case "find":
						newToken := estructuras.NewToken("TK_FIND", lexemaact)
						tokens = append(tokens, newToken)
					case "chown":
						newToken := estructuras.NewToken("TK_CHOWN", lexemaact)
						tokens = append(tokens, newToken)
					case "chgrp":
						newToken := estructuras.NewToken("TK_CHGRP", lexemaact)
						tokens = append(tokens, newToken)
					case "loss":
						newToken := estructuras.NewToken("TK_LOSS", lexemaact)
						tokens = append(tokens, newToken)
					case "recovery":
						newToken := estructuras.NewToken("TK_RECOVERY", lexemaact)
						tokens = append(tokens, newToken)
					default:
						//color.Println(lexemaact)
						newToken := estructuras.NewToken("TK_ID", lexemaact)
						tokens = append(tokens, newToken)
					}
					idee = false
					estado = 0
					if !anular {
						repetir = true
					} else if anular {
						repetir = false
						anular = false
					}
					lexemaact = ""
				}

			default:
				fmt.Println("")
			}
		}
	}
	if !errorLex {
		Sintactico()
	} else {
		fmt.Println("Error Lexico encontrado")
	}
}

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func isLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isWhiteSpace(s string) bool {
	switch s {
	case " ", "\t", "\n", "\f", "\r":
		return true
	}
	return false
}
