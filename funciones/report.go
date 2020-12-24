package funciones

import (
	"Proyecto1/estructuras"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/doun/terminal/color"
)

var (
	contadorAVD, contadorDD, contadorInodo, contadorBloque, contadorBitacoras, cont int    = 0, 0, 0, 0, 0, 0
	cadenaArchivo                                                                   string = ""
)

//EjecutarReporte verifica el tipo de reporte segun el parametro NOMBRE
func EjecutarReporte(nombre string, path string, ruta string, id string) {
	color.Println(nombre + "n " + path + "p " + ruta + "r " + id)
	if nombre == "" {
		nombre = string(cont)
	}
	if nombre != "" && id != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil { //verificamos que se pueda construir el path
			fmt.Printf("Path invalido")
		} else {

			if IDYaRegistrado(id) { //verificamos que el id si exista, osea que haya una particion montada con ese id
				if strings.ToLower(nombre) == "mbr" {
					ReporteMBR(path, ruta, id)
				} else if strings.ToLower(nombre) == "disk" {
					ReporteDisk(path, ruta, id)
				} else if strings.ToLower(nombre) == "bm_arbdir" {
					ReporteBitmapAVD(path, ruta, id)
				} else if strings.ToLower(nombre) == "bm_detdir" {
					ReporteBitmapDD(path, ruta, id)
				} else if strings.ToLower(nombre) == "bm_inode" {
					ReporteBitmapInode(path, ruta, id)
				} else if strings.ToLower(nombre) == "bm_block" {
					ReporteBitmapBloque(path, ruta, id)
				} else if strings.ToLower(nombre) == "sb" {
					ReporteSB(path, ruta, id)
				} else if strings.ToLower(nombre) == "tree" {
					ReporteTreeComplete(path, ruta, id)
				} else if strings.ToLower(nombre) == "inode" {
					ReporteTreeComplete3(path, ruta, id)
				} else if strings.ToLower(nombre) == "block" {
					ReporteTreeComplete4(path, ruta, id)
				} else if strings.ToLower(nombre) == "journaling" {
					ReporteBitacora(path, ruta, id)
				} else if strings.ToLower(nombre) == "ls" {
					EjecutarLS(ruta, id)
				} else if strings.ToLower(nombre) == "blockes" {
					ReporteTreeDirectorio(path, "/", id)
				}
			} else {
				color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
			}
		}
	} else {
		color.Println("@{r}Faltan parámetros obligatorios para la funcion REP.")
	}

}

//ReporteMBR crea el reporte del mbr
func ReporteMBR(path string, ruta string, id string) {

	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		file, err := os.OpenFile("codigo3.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}
		// Change permissions Linux.
		err = os.Chmod("codigo3.dot", 0666)
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}

		file.Truncate(0)
		file.Seek(0, 0)

		_, err = file.WriteString("digraph H {\n node [ shape=plain] \n table [ label = <\n  <table border='1' cellborder='1'>\n   <tr><td>Nombre</td><td>Valor</td></tr>\n")
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}

		//LEER Y RECORRER EL MBR
		_, PathAux := GetDatosPart(id)
		fileMBR, err2 := os.Open(PathAux)
		if err2 != nil { //validar que no sea nulo.
			panic(err2)
		}
		Disco1 := estructuras.MBR{}
		DiskSize := int(unsafe.Sizeof(Disco1))
		DiskData := leerBytes(fileMBR, DiskSize)
		buffer := bytes.NewBuffer(DiskData)
		err = binary.Read(buffer, binary.BigEndian, &Disco1)
		if err != nil {
			fileMBR.Close()
			fmt.Println(err)
			return
		}
		fileMBR.Close()

		w := bufio.NewWriter(file)

		fmt.Fprintf(w, "   <tr><td>MBR_Tamanio</td><td>%v</td></tr>\n", Disco1.Msize)
		fmt.Fprintf(w, "   <tr><td>MBR_Fecha_Creación</td><td>%v</td></tr>\n", string(Disco1.Mdate[:len(Disco1.Mdate)-1]))
		fmt.Fprintf(w, "   <tr><td>MBR_Disk_Signature</td><td>%v</td></tr>\n", Disco1.Msignature)

		PartNum := 1
		for i := 0; i < 4; i++ {
			if Disco1.Mpartitions[i].Psize > 0 {
				fmt.Fprintf(w, "   <tr><td>Part_%d_Status</td><td>%v</td></tr>\n", PartNum, string(Disco1.Mpartitions[i].Pstatus))
				fmt.Fprintf(w, "   <tr><td>Part_%d_Type</td><td>%v</td></tr>\n", PartNum, string(Disco1.Mpartitions[i].Ptype))
				fmt.Fprintf(w, "   <tr><td>Part_%d_Fit</td><td>%v</td></tr>\n", PartNum, string(Disco1.Mpartitions[i].Pfit))
				fmt.Fprintf(w, "   <tr><td>Part_%d_Start</td><td>%v</td></tr>\n", PartNum, Disco1.Mpartitions[i].Pstart)
				n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
				if n == -1 {
					n = len(Disco1.Mpartitions[i].Pname)
				}
				fmt.Fprintf(w, "   <tr><td>Part_%d_Name</td><td>%v</td></tr>\n", PartNum, string(Disco1.Mpartitions[i].Pname[:n]))
				PartNum++
			}
		}

		w.Flush()
		////////////////////
		_, err = file.WriteString("  </table>\n > ]\n}")
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}

		file.Close()

		extT := "-T"

		switch strings.ToLower(extension) {
		case ".png":
			extT += "png"
		case ".pdf":
			extT += "pdf"
		case ".jpg":
			extT += "jpg"
		default:

		}

		if runtime.GOOS == "windows" {
			cmd := exec.Command("dot", extT, "-o", path, "codigo3.dot") //Windows example, its tested
			//cmd.Stdout = os.Stdout
			cmd.Run()
		} else {
			cmd := exec.Command("dot", extT, "-o", path, "codigo3.dot") //Linux example, its tested
			//cmd.Stdout = os.Stdout
			cmd.Run()
		}

		color.Print("@{w}El reporte@{w} MBR @{w}fue creado con éxito\n")

	} else {
		color.Println("@{r}El reporte MBR solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}

}

//ReporteDisk crea el reporte de las particiones del disco
func ReporteDisk(path string, ruta string, id string) {

	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		file, err := os.OpenFile("codigo4.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}
		// Change permissions Linux.
		err = os.Chmod("codigo4.dot", 0666)
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}

		file.Truncate(0)
		file.Seek(0, 0)

		_, err = file.WriteString("digraph D {\nparent [\n shape=plaintext\n label=<\n<table border='1' cellborder='1'>\n <tr>\n  <td rowspan=\"2\" bgcolor=\"orange\">MBR</td>\n")
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}

		//LEER Y RECORRER EL MBR
		_, PathAux := GetDatosPart(id)
		fileMBR, err2 := os.Open(PathAux)
		if err2 != nil { //validar que no sea nulo.
			panic(err2)
		}
		Disco1 := estructuras.MBR{}
		DiskSize := int(unsafe.Sizeof(Disco1))
		DiskData := leerBytes(fileMBR, DiskSize)
		buffer := bytes.NewBuffer(DiskData)
		err = binary.Read(buffer, binary.BigEndian, &Disco1)
		if err != nil {
			fileMBR.Close()
			fmt.Println(err)
			return
		}
		fileMBR.Close()

		w := bufio.NewWriter(file)
		HayExtendida := false
		nLogicas := 0
		for i := 0; i < 4; i++ {
			if Disco1.Mpartitions[i].Psize > 0 {

				if Disco1.Mpartitions[i].Ptype == 'e' || Disco1.Mpartitions[i].Ptype == 'E' {
					HayExtendida = true
					nLogicas = CantidadLogicas(PathAux)
					Columnas := 0
					if nLogicas > 0 {
						Columnas = 2 * nLogicas
					} else {
						Columnas = 2
					}

					if i > 0 {

						if Disco1.Mpartitions[i-1].Psize == 0 {
							n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
							if n == -1 {
								n = len(Disco1.Mpartitions[i].Pname)
							}
							fmt.Fprintf(w, "  <td colspan=\"%d\" bgcolor=\"#0cff04\">%v - Extendida</td>\n", Columnas, string(Disco1.Mpartitions[i].Pname[:n]))
						} else {

							if (Disco1.Mpartitions[i-1].Pstart + Disco1.Mpartitions[i-1].Psize) == Disco1.Mpartitions[i].Pstart {
								n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
								if n == -1 {
									n = len(Disco1.Mpartitions[i].Pname)
								}
								fmt.Fprintf(w, "  <td colspan=\"%d\" bgcolor=\"#0cff04\">%v - Extendida</td>\n", Columnas, string(Disco1.Mpartitions[i].Pname[:n]))
							} else {
								fmt.Fprint(w, "  <td rowspan=\"2\" bgcolor=\"skyblue\">Libre</td>")
								n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
								if n == -1 {
									n = len(Disco1.Mpartitions[i].Pname)
								}
								fmt.Fprintf(w, "  <td colspan=\"%d\" bgcolor=\"#0cff04\">%v - Extendida</td>\n", Columnas, string(Disco1.Mpartitions[i].Pname[:n]))
							}

						}

					} else {
						n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
						fmt.Fprintf(w, "  <td colspan=\"%d\" bgcolor=\"#0cff04\">%v - Extendida</td>\n", Columnas, string(Disco1.Mpartitions[i].Pname[:n]))
					}

				} else {

					if i > 0 {

						if Disco1.Mpartitions[i-1].Psize == 0 {
							n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
							if n == -1 {
								n = len(Disco1.Mpartitions[i].Pname)
							}
							fmt.Fprintf(w, "  <td rowspan=\"2\" bgcolor=\"yellow\">%v - Primaria</td>\n", string(Disco1.Mpartitions[i].Pname[:n]))
						} else {

							if (Disco1.Mpartitions[i-1].Pstart + Disco1.Mpartitions[i-1].Psize) == Disco1.Mpartitions[i].Pstart {
								n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
								if n == -1 {
									n = len(Disco1.Mpartitions[i].Pname)
								}
								fmt.Fprintf(w, "  <td rowspan=\"2\" bgcolor=\"yellow\">%v - Primaria</td>\n", string(Disco1.Mpartitions[i].Pname[:n]))
							} else {
								fmt.Fprint(w, "  <td rowspan=\"2\" bgcolor=\"skyblue\">Libre</td>")
								n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
								if n == -1 {
									n = len(Disco1.Mpartitions[i].Pname)
								}
								fmt.Fprintf(w, "  <td rowspan=\"2\" bgcolor=\"yellow\">%v - Primaria</td>\n", string(Disco1.Mpartitions[i].Pname[:n]))
							}
						}

					} else {
						n := bytes.Index(Disco1.Mpartitions[i].Pname[:], []byte{0})
						if n == -1 {
							n = len(Disco1.Mpartitions[i].Pname)
						}
						fmt.Fprintf(w, "  <td rowspan=\"2\" bgcolor=\"yellow\">%v - Primaria</td>\n", string(Disco1.Mpartitions[i].Pname[:n]))
					}
				}

				if i == 3 {
					if Disco1.Mpartitions[i].Pstart+Disco1.Mpartitions[i].Psize != Disco1.Msize-1 {
						fmt.Fprint(w, "  <td rowspan=\"2\" bgcolor=\"skyblue\">Libre</td>\n")
					}

				}

			} else {

				if i > 0 {
					if Disco1.Mpartitions[i-1].Psize != 0 {
						fmt.Fprint(w, "  <td rowspan=\"2\" bgcolor=\"skyblue\">Libre</td>\n")
					}
				} else {
					fmt.Fprint(w, "  <td rowspan=\"2\" bgcolor=\"skyblue\">Libre</td>\n")
				}

			}
		}
		fmt.Fprint(w, " </tr>\n")

		if HayExtendida {
			fmt.Fprint(w, " <tr>\n")

			////////////////////RECORRER EBRS

			indiceExt, _ := IndiceExtendida(PathAux)

			fileEBR, err := os.OpenFile(PathAux, os.O_RDWR, 0666)
			if err != nil {
				fmt.Println(err)
				fileEBR.Close()
			}

			EBRaux := estructuras.EBR{}
			EBRSize := int(unsafe.Sizeof(EBRaux))
			fileEBR.Seek(int64(indiceExt)+1, 0)
			EBRData := leerBytes(fileEBR, EBRSize)
			buffer := bytes.NewBuffer(EBRData)

			err = binary.Read(buffer, binary.BigEndian, &EBRaux)
			if err != nil {
				fileEBR.Close()
				panic(err)
			}

			Continuar := true

			for Continuar {

				if EBRaux.Esize > 0 {
					fmt.Fprint(w, "  <td>EBR</td>\n")
					fmt.Fprint(w, "  <td>LOGICA</td>\n")
				} else {
					fmt.Fprint(w, "  <td>EBR</td>\n")
					fmt.Fprint(w, "  <td>LIBRE</td>\n")
				}

				if EBRaux.Enext != -1 {
					//Si hay otro EBR a la derecha lo leemos y volvemos al inicio del FOR
					fileEBR.Seek(int64(EBRaux.Enext)+1, 0)
					EBRData := leerBytes(fileEBR, EBRSize)
					buffer := bytes.NewBuffer(EBRData)
					err = binary.Read(buffer, binary.BigEndian, &EBRaux)
					if err != nil {
						fileEBR.Close()
						panic(err)
					}
				} else {
					//Si no cancelamos el loop
					Continuar = false
				}

			}
			fileEBR.Close()

			////////////////////////////////////
			fmt.Fprint(w, " </tr>\n")
		}

		w.Flush()

		////////////////////
		_, err = file.WriteString("  </table>\n >]\n}")
		if err != nil {
			fmt.Println(err)
			file.Close()
			return
		}

		file.Close()

		extT := "-T"

		switch strings.ToLower(extension) {
		case ".png":
			extT += "png"
		case ".pdf":
			extT += "pdf"
		case ".jpg":
			extT += "jpg"
		default:

		}

		if runtime.GOOS == "windows" {
			cmd := exec.Command("dot", extT, "-o", path, "codigo4.dot") //Windows example, its tested
			//cmd.Stdout = os.Stdout
			cmd.Run()
		} else {
			cmd := exec.Command("dot", extT, "-o", path, "codigo4.dot") //Linux example, its tested
			//cmd.Stdout = os.Stdout
			cmd.Run()
		}

		color.Print("@{w}El reporte@{w} DISK @{w}fue creado con éxito\n")

	} else {
		color.Println("@{r}El reporte DISK solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}
}

//ReporteSB crea el reporte del super bloque
func ReporteSB(path string, ruta string, id string) {

	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}
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
			fileMBR.Close()

			if SB1.MontajesCount > 0 {

				file, err := os.OpenFile("codigo6.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod("codigo6.dot", 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)

				_, err = file.WriteString("digraph H {\n node [ shape=plain] \n table [ label = <\n  <table border='1' cellborder='1'>\n   <tr><td>Atributo</td><td>Valor</td></tr>\n")
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				w := bufio.NewWriter(file)

				///////////SETEAR DATOS DEL SUPERBLOQUE
				n := bytes.Index(SB1.Name[:], []byte{0})
				if n == -1 {
					n = len(SB1.Name)
				}
				fmt.Fprintf(w, "   <tr><td>s_filesystem_type</td><td>%v</td></tr>\n", "3")
				//fmt.Fprintf(w, "   <tr><td>Total AVDs</td><td>%v</td></tr>\n", SB1.TotalAVDS)
				//fmt.Fprintf(w, "   <tr><td>Total DDs</td><td>%v</td></tr>\n", SB1.TotalDDS)
				fmt.Fprintf(w, "   <tr><td>s_inodes_count</td><td>%v</td></tr>\n", SB1.TotalInodos)
				fmt.Fprintf(w, "   <tr><td>s_blocks_count</td><td>%v</td></tr>\n", SB1.TotalBloques)
				//fmt.Fprintf(w, "   <tr><td>Total Bitacoras</td><td>%v</td></tr>\n", SB1.TotalBitacoras)

				//fmt.Fprintf(w, "   <tr><td>Contador AVDs</td><td>%v</td></tr>\n", int(SB1.TotalAVDS-SB1.FreeAVDS))
				//fmt.Fprintf(w, "   <tr><td>Contador DDs</td><td>%v</td></tr>\n", int(SB1.TotalDDS-SB1.FreeDDS))
				fmt.Fprintf(w, "   <tr><td>s_Inodos_usados</td><td>%v</td></tr>\n", int(SB1.TotalInodos-SB1.FreeInodos))
				fmt.Fprintf(w, "   <tr><td>s_Bloques_usados</td><td>%v</td></tr>\n", int(SB1.TotalBloques-SB1.FreeBloques))
				//fmt.Fprintf(w, "   <tr><td>Contador Bitacoras</td><td>%v</td></tr>\n", int(SB1.TotalBitacoras-SB1.FreeBitacoras))

				//	fmt.Fprintf(w, "   <tr><td>Free AVDs</td><td>%v</td></tr>\n", SB1.FreeAVDS)
				//	fmt.Fprintf(w, "   <tr><td>Free DDs</td><td>%v</td></tr>\n", SB1.FreeDDS)
				fmt.Fprintf(w, "   <tr><td>s_Free_Inodos</td><td>%v</td></tr>\n", SB1.FreeInodos)
				fmt.Fprintf(w, "   <tr><td>s_Free_Bloques</td><td>%v</td></tr>\n", SB1.FreeBloques)
				//fmt.Fprintf(w, "   <tr><td>Free Bitacoras</td><td>%v</td></tr>\n", SB1.FreeBitacoras)
				n = bytes.Index(SB1.DateCreacion[:], []byte{0})
				if n == -1 {
					n = len(SB1.DateCreacion)
				}
				fmt.Fprintf(w, "   <tr><td>s_mtime</td><td>%v</td></tr>\n", string(SB1.DateCreacion[:n]))
				n = bytes.Index(SB1.DateLastMount[:], []byte{0})
				if n == -1 {
					n = len(SB1.DateLastMount)
				}
				fmt.Fprintf(w, "   <tr><td>Fecha montaje</td><td>%v</td></tr>\n", string(SB1.DateLastMount[:n]))
				fmt.Fprintf(w, "   <tr><td>s_mnt_count</td><td>%v</td></tr>\n", SB1.MontajesCount)
				fmt.Fprintf(w, "   <tr><td>s_Magic</td><td>%v</td></tr>\n", "0xEF53")
				//fmt.Fprintf(w, "   <tr><td>Apuntador Bitmap AVDs</td><td>%v</td></tr>\n", SB1.InicioBitmapAVDS)
				//fmt.Fprintf(w, "   <tr><td>Apuntador AVDs</td><td>%v</td></tr>\n", SB1.InicioAVDS)
				//fmt.Fprintf(w, "   <tr><td>Apuntador Bitmap DDs</td><td>%v</td></tr>\n", SB1.InicioBitMapDDS)
				//fmt.Fprintf(w, "   <tr><td>Apuntador DDs</td><td>%v</td></tr>\n", SB1.InicioDDS)

				//fmt.Fprintf(w, "   <tr><td>Apuntador Bitacoras</td><td>%v</td></tr>\n", SB1.InicioBitacora)
				//fmt.Fprintf(w, "   <tr><td>Size struct AVD</td><td>%v</td></tr>\n", SB1.SizeAVD)
				//fmt.Fprintf(w, "   <tr><td>Size struct DD</td><td>%v</td></tr>\n", SB1.SizeDD)
				fmt.Fprintf(w, "   <tr><td>S_inode_size</td><td>%v</td></tr>\n", SB1.SizeInodo)
				fmt.Fprintf(w, "   <tr><td>s_block_size</td><td>%v</td></tr>\n", SB1.SizeBloque)
				//fmt.Fprintf(w, "   <tr><td>Byte primer AVD libre</td><td>%v</td></tr>\n", SB1.FirstFreeAVD)
				//fmt.Fprintf(w, "   <tr><td>Byte primer DD libre</td><td>%v</td></tr>\n", SB1.FirstFreeDD)
				fmt.Fprintf(w, "   <tr><td>s_first_ino</td><td>%v</td></tr>\n", SB1.FirstFreeInodo)
				fmt.Fprintf(w, "   <tr><td>s_first_blo</td><td>%v</td></tr>\n", SB1.FirstFreeBloque)
				fmt.Fprintf(w, "   <tr><td>s_bm_inode_start</td><td>%v</td></tr>\n", SB1.InicioBitmapInodos)
				fmt.Fprintf(w, "   <tr><td>s_inode_start</td><td>%v</td></tr>\n", SB1.InicioInodos)
				fmt.Fprintf(w, "   <tr><td>s_bm_block_start</td><td>%v</td></tr>\n", SB1.InicioBitmapBloques)
				fmt.Fprintf(w, "   <tr><td>s_block_start</td><td>%v</td></tr>\n", SB1.InicioBloques)
				////////////////////

				w.Flush()

				_, err = file.WriteString("  </table>\n > ]\n}")
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Close()

				extT := "-T"

				switch strings.ToLower(extension) {
				case ".png":
					extT += "png"
				case ".pdf":
					extT += "pdf"
				case ".jpg":
					extT += "jpg"
				default:

				}

				if runtime.GOOS == "windows" {
					cmd := exec.Command("dot", extT, "-o", path, "codigo6.dot") //Windows example, its tested
					//cmd.Stdout = os.Stdout
					cmd.Run()
				} else {
					cmd := exec.Command("dot", extT, "-o", path, "codigo6.dot") //Linux example, its tested
					//cmd.Stdout = os.Stdout
					cmd.Run()
				}

				color.Print("@{w}El reporte@{w} SB @{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

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
			fileMBR.Close()
			if SB1.MontajesCount > 0 {

				file, err := os.OpenFile("codigo6.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod("codigo6.dot", 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)

				_, err = file.WriteString("digraph H {\n node [ shape=plain] \n table [ label = <\n  <table border='1' cellborder='1'>\n   <tr><td>Atributo</td><td>Valor</td></tr>\n")
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				w := bufio.NewWriter(file)

				///////////SETEAR DATOS DEL SUPERBLOQUE
				n := bytes.Index(SB1.Name[:], []byte{0})
				fmt.Fprintf(w, "   <tr><td>s_filesystem_type</td><td>%v</td></tr>\n", "3")
				//fmt.Fprintf(w, "   <tr><td>Total AVDs</td><td>%v</td></tr>\n", SB1.TotalAVDS)
				//fmt.Fprintf(w, "   <tr><td>Total DDs</td><td>%v</td></tr>\n", SB1.TotalDDS)
				fmt.Fprintf(w, "   <tr><td>s_inodes_count</td><td>%v</td></tr>\n", SB1.TotalInodos)
				fmt.Fprintf(w, "   <tr><td>s_block_count</td><td>%v</td></tr>\n", SB1.TotalBloques)
				//	fmt.Fprintf(w, "   <tr><td>Total Bitacoras</td><td>%v</td></tr>\n", SB1.TotalBitacoras)

				//	fmt.Fprintf(w, "   <tr><td>Contador AVDs</td><td>%v</td></tr>\n", int(SB1.TotalAVDS-SB1.FreeAVDS))
				//	fmt.Fprintf(w, "   <tr><td>Contador DDs</td><td>%v</td></tr>\n", int(SB1.TotalDDS-SB1.FreeDDS))
				fmt.Fprintf(w, "   <tr><td>s_inodes_usados</td><td>%v</td></tr>\n", int(SB1.TotalInodos-SB1.FreeInodos))
				fmt.Fprintf(w, "   <tr><td>s_block_usados</td><td>%v</td></tr>\n", int(SB1.TotalBloques-SB1.FreeBloques))
				//	fmt.Fprintf(w, "   <tr><td>Contador Bitacoras</td><td>%v</td></tr>\n", int(SB1.TotalBitacoras-SB1.FreeBitacoras))

				//	fmt.Fprintf(w, "   <tr><td>Free AVDs</td><td>%v</td></tr>\n", SB1.FreeAVDS)
				//	fmt.Fprintf(w, "   <tr><td>Free DDs</td><td>%v</td></tr>\n", SB1.FreeDDS)
				fmt.Fprintf(w, "   <tr><td>s_Free_Inodos</td><td>%v</td></tr>\n", SB1.FreeInodos)
				fmt.Fprintf(w, "   <tr><td>s_Free_Bloques</td><td>%v</td></tr>\n", SB1.FreeBloques)
				//	fmt.Fprintf(w, "   <tr><td>Free Bitacoras</td><td>%v</td></tr>\n", SB1.FreeBitacoras)
				n = bytes.Index(SB1.DateCreacion[:], []byte{0})
				if n == -1 {
					n = len(SB1.DateCreacion)
				}
				fmt.Fprintf(w, "   <tr><td>Fecha creación</td><td>%v</td></tr>\n", string(SB1.DateCreacion[:n]))
				n = bytes.Index(SB1.DateLastMount[:], []byte{0})
				if n == -1 {
					n = len(SB1.DateLastMount)
				}
				fmt.Fprintf(w, "   <tr><td>Fecha Modificación</td><td>%v</td></tr>\n", string(SB1.DateLastMount[:n]))
				fmt.Fprintf(w, "   <tr><td>No. Montajes</td><td>%v</td></tr>\n", SB1.MontajesCount)
				//	fmt.Fprintf(w, "   <tr><td>Apuntador Bitmap AVDs</td><td>%v</td></tr>\n", SB1.InicioBitmapAVDS)
				//	fmt.Fprintf(w, "   <tr><td>Apuntador AVDs</td><td>%v</td></tr>\n", SB1.InicioAVDS)
				//	fmt.Fprintf(w, "   <tr><td>Apuntador Bitmap DDs</td><td>%v</td></tr>\n", SB1.InicioBitMapDDS)
				//	fmt.Fprintf(w, "   <tr><td>Apuntador DDs</td><td>%v</td></tr>\n", SB1.InicioDDS)
				fmt.Fprintf(w, "   <tr><td>Apuntador Bitmap Inodos</td><td>%v</td></tr>\n", SB1.InicioBitmapInodos)
				fmt.Fprintf(w, "   <tr><td>Apuntador inodos</td><td>%v</td></tr>\n", SB1.InicioInodos)
				fmt.Fprintf(w, "   <tr><td>Apuntador Bitmap bloques</td><td>%v</td></tr>\n", SB1.InicioBitmapBloques)
				fmt.Fprintf(w, "   <tr><td>Apuntador Bloques</td><td>%v</td></tr>\n", SB1.InicioBloques)
				//	fmt.Fprintf(w, "   <tr><td>Apuntador Bitacoras</td><td>%v</td></tr>\n", SB1.InicioBitacora)
				//	fmt.Fprintf(w, "   <tr><td>Size struct AVD</td><td>%v</td></tr>\n", SB1.SizeAVD)
				//	fmt.Fprintf(w, "   <tr><td>Size struct DD</td><td>%v</td></tr>\n", SB1.SizeDD)
				fmt.Fprintf(w, "   <tr><td>Size struct Inodo</td><td>%v</td></tr>\n", SB1.SizeInodo)
				fmt.Fprintf(w, "   <tr><td>Size struct Bloque</td><td>%v</td></tr>\n", SB1.SizeBloque)
				//	fmt.Fprintf(w, "   <tr><td>Byte primer AVD libre</td><td>%v</td></tr>\n", SB1.FirstFreeAVD)
				//	fmt.Fprintf(w, "   <tr><td>Byte primer DD libre</td><td>%v</td></tr>\n", SB1.FirstFreeDD)
				fmt.Fprintf(w, "   <tr><td>Byte primer Inodo libre</td><td>%v</td></tr>\n", SB1.FirstFreeInodo)
				fmt.Fprintf(w, "   <tr><td>Byte primer Bloque libre</td><td>%v</td></tr>\n", SB1.FirstFreeBloque)
				fmt.Fprintf(w, "   <tr><td>Magic Number :D </td><td>%v</td></tr>\n", SB1.MagicNum)
				////////////////////

				w.Flush()

				_, err = file.WriteString("  </table>\n > ]\n}")
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Close()

				extT := "-T"

				switch strings.ToLower(extension) {
				case ".png":
					extT += "png"
				case ".pdf":
					extT += "pdf"
				case ".jpg":
					extT += "jpg"
				default:

				}

				if runtime.GOOS == "windows" {
					cmd := exec.Command("dot", extT, "-o", path, "codigo6.dot") //Windows example, its tested
					//cmd.Stdout = os.Stdout
					cmd.Run()
				} else {
					cmd := exec.Command("dot", extT, "-o", path, "codigo6.dot") //Linux example, its tested
					//cmd.Stdout = os.Stdout
					cmd.Run()
				}

				color.Print("@{w}El reporte@{w} SB @{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

		}

	} else {
		color.Println("@{r}El reporte SB solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}

}

//ReporteBitacora crea el reporte de las bitacoras
func ReporteBitacora(path string, ruta string, id string) {

	extension := filepath.Ext(path)
	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}
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

				NumeroBitacoras := int(SB1.TotalBitacoras - SB1.FreeBitacoras)

				if NumeroBitacoras > 0 {

					file, err := os.OpenFile("codigo9.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
					if err != nil {
						fmt.Println(err)
						file.Close()
						return
					}

					// Change permissions Linux.
					err = os.Chmod("codigo9.dot", 0666)
					if err != nil {
						fmt.Println(err)
						file.Close()
						return
					}

					file.Truncate(0)
					file.Seek(0, 0)

					w := bufio.NewWriter(file)

					fmt.Fprint(w, `digraph Journaling {
					node [shape=plaintext];
					`)

					///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

					contadorBitacoras = 0
					cadenaArchivo = ""

					for i := 0; i < NumeroBitacoras; i++ {

						//LEER EL SUPERBLOQUE
						InicioBitacora := SB1.InicioBitacora + (int32(i) * SB1.SizeBitacora)
						fileMBR.Seek(int64(InicioBitacora+1), 0)
						BitacoraAux := estructuras.Bitacora{}
						BitacoraSize := int(unsafe.Sizeof(BitacoraAux))
						BitacoraData := leerBytes(fileMBR, BitacoraSize)
						buffer := bytes.NewBuffer(BitacoraData)
						err = binary.Read(buffer, binary.BigEndian, &BitacoraAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return
						}

						cadenaArchivo += GenerarBitacora(contadorBitacoras, &BitacoraAux)
						contadorBitacoras++

						if (i + 1) < NumeroBitacoras {
							cadenaArchivo += fmt.Sprintf(`B%v->B%v[style=invis]
							
							`, contadorBitacoras-1, contadorBitacoras)
						}

					}

					fmt.Fprint(w, cadenaArchivo)

					///////////////////////////////////////////////////////////////////////////////////////

					fmt.Fprint(w, `}`)

					w.Flush()

					file.Close()

					extT := "-T"

					switch strings.ToLower(extension) {
					case ".png":
						extT += "png"
					case ".pdf":
						extT += "pdf"
					case ".jpg":
						extT += "jpg"
					default:

					}

					carpeta := filepath.Dir(path)
					SVGpath := carpeta + "/Journaling.svg"

					if runtime.GOOS == "windows" {
						cmd := exec.Command("dot", extT, "-o", path, "codigo9.dot")
						cmd.Run()
						cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo9.dot")
						cmd2.Run()
					} else {
						cmd := exec.Command("dot", extT, "-o", path, "codigo9.dot")
						cmd.Run()
						cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo9.dot")
						cmd2.Run()
					}

					color.Print("@{w}El reporte@{w} Journaling @{w}fue creado con éxito\n")

				} else {
					color.Println("@{r} No hay journalong en el sistema.")
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

				NumeroBitacoras := int(SB1.TotalBitacoras - SB1.FreeBitacoras)

				if NumeroBitacoras > 0 {

					file, err := os.OpenFile("codigo9.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
					if err != nil {
						fmt.Println(err)
						file.Close()
						return
					}

					// Change permissions Linux.
					err = os.Chmod("codigo9.dot", 0666)
					if err != nil {
						fmt.Println(err)
						file.Close()
						return
					}

					file.Truncate(0)
					file.Seek(0, 0)

					w := bufio.NewWriter(file)

					fmt.Fprint(w, `digraph Journaling {
					node [shape=plaintext];
					`)

					///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

					contadorBitacoras = 0
					cadenaArchivo = ""

					for i := 0; i < NumeroBitacoras; i++ {

						//LEER EL SUPERBLOQUE
						InicioBitacora := SB1.InicioBitacora + (int32(i) * SB1.SizeBitacora)
						fileMBR.Seek(int64(InicioBitacora+1), 0)
						BitacoraAux := estructuras.Bitacora{}
						BitacoraSize := int(unsafe.Sizeof(BitacoraAux))
						BitacoraData := leerBytes(fileMBR, BitacoraSize)
						buffer := bytes.NewBuffer(BitacoraData)
						err = binary.Read(buffer, binary.BigEndian, &BitacoraAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return
						}

						cadenaArchivo += GenerarBitacora(contadorBitacoras, &BitacoraAux)
						contadorBitacoras++

						if (i + 1) < NumeroBitacoras {
							cadenaArchivo += fmt.Sprintf(`B%v->B%v[style=invis]
							
							`, contadorBitacoras-1, contadorBitacoras)
						}

					}

					fmt.Fprint(w, cadenaArchivo)

					///////////////////////////////////////////////////////////////////////////////////////

					fmt.Fprint(w, `}`)

					w.Flush()

					file.Close()

					extT := "-T"

					switch strings.ToLower(extension) {
					case ".png":
						extT += "png"
					case ".pdf":
						extT += "pdf"
					case ".jpg":
						extT += "jpg"
					default:

					}

					carpeta := filepath.Dir(path)
					SVGpath := carpeta + "/Journaling.svg"

					if runtime.GOOS == "windows" {
						cmd := exec.Command("dot", extT, "-o", path, "codigo9.dot")
						cmd.Run()
						cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo9.dot")
						cmd2.Run()
					} else {
						cmd := exec.Command("dot", extT, "-o", path, "codigo9.dot")
						cmd.Run()
						cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo9.dot")
						cmd2.Run()
					}

					color.Print("@{w}El reporte@{w} Journaling @{w}fue creado con éxito\n")

				} else {
					color.Println("@{r} No hay jornaling en el sistema.")
				}

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()

		}

	} else {
		color.Println("@{r}El reporte Journaling A solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}

}

//ReporteTreeComplete crea el reporte del sistema completo Generarbloque
func ReporteTreeComplete(path string, ruta string, id string) {

	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}

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

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDroot := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDroot))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				RootData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(RootData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDroot)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				EscribirTreeComplete(fileMBR, &AVDroot, extension, path)

				color.Print("@{w}El reporte@{w} TREE @{w}fue creado con éxito\n")

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

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDroot := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDroot))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				RootData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(RootData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDroot)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				EscribirTreeComplete(fileMBR, &AVDroot, extension, path)

				color.Print("@{w}El reporte@{w} TREE @{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte TREE_COMPLETE solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}

}

//EscribirTreeComplete genera el reporte TreeComplete al recibir la AVD de la raiz
func EscribirTreeComplete(MBRfile *os.File, AVDroot *estructuras.AVD, extension string, path string) {

	file, err := os.OpenFile("codigo7.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	// Change permissions Linux.
	err = os.Chmod("codigo7.dot", 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	file.Truncate(0)
	file.Seek(0, 0)

	w := bufio.NewWriter(file)

	fmt.Fprint(w, `digraph Tree {
		node [shape=plaintext];
		`)

	///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

	contadorAVD = 0
	contadorDD = 0
	contadorInodo = 0
	contadorBloque = 0
	cadenaArchivo = ""

	EscribirAVDRecursivo(MBRfile, AVDroot, contadorAVD)

	fmt.Fprint(w, cadenaArchivo)

	///////////////////////////////////////////////////////////////////////////////////////

	fmt.Fprint(w, `}`)

	w.Flush()

	file.Close()

	extT := "-T"

	switch strings.ToLower(extension) {
	case ".png":
		extT += "png"
	case ".pdf":
		extT += "pdf"
	case ".jpg":
		extT += "jpg"
	default:

	}

	carpeta := filepath.Dir(path)
	SVGpath := carpeta + "/tree.svg"

	if runtime.GOOS == "windows" {
		cmd := exec.Command("dot", extT, "-o", path, "codigo7.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo7.dot")
		cmd2.Run()
	} else {
		cmd := exec.Command("dot", extT, "-o", path, "codigo7.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo7.dot")
		cmd2.Run()
	}

}

//EscribirAVDRecursivo recorre un AVD
func EscribirAVDRecursivo(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	cadenaArchivo += GenerarAVD(contadorAVD, AVDAux)
	contadorAVD++

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

			cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v
			
				`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo(file, &AVDHijo, contadorAVD)

		}
	}

	cadenaArchivo += fmt.Sprintf(`AVD%v:6->DD%v

			`, NoAVD, contadorDD)

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

	n := bytes.Index(AVDAux.NombreDir[:], []byte{0})
	if n == -1 {
		n = len(AVDAux.NombreDir)
	}
	carpeta := string(AVDAux.NombreDir[:n])
	EscribirDDRecursivo(file, &DDAux, contadorDD, carpeta)

	if AVDAux.ApuntadorAVD > 0 {

		cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

			`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo(file, &AVDExt, contadorAVD)
	}

}

//EscribirAVDExtRecursivo recorre la extensión del AVD
func EscribirAVDExtRecursivo(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	cadenaArchivo += GenerarAVD(contadorAVD, AVDAux)
	contadorAVD++

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

			cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v
			
				`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo(file, &AVDHijo, contadorAVD)

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

		cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v
	
		`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo(file, &AVDExt, contadorAVD)
	}

}

//EscribirDDRecursivo recorre el detalle de directorio
func EscribirDDRecursivo(file *os.File, DDaux *estructuras.DD, NoDD int, carpeta string) {

	cadenaArchivo += GenerarDD(NoDD, DDaux, carpeta)
	contadorDD++

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

			cadenaArchivo += fmt.Sprintf(`DD%v:%v->Inodo%v
			
				`, NoDD, i, contadorInodo)

			EscribirInodoRecursivo(file, &InodoAux, contadorInodo)

		}
	}

	if DDaux.ApuntadorDD > 0 {

		cadenaArchivo += fmt.Sprintf(`DD%v:5->DD%v
	
		`, NoDD, contadorDD)

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

		EscribirDDRecursivo(file, &DDExt, contadorDD, carpeta)
	}

}

//EscribirInodoRecursivo recorre el inodo generarInodo
func EscribirInodoRecursivo(file *os.File, InodoAux *estructuras.Inodo, NoInodo int) {

	cadenaArchivo += GenerarInodo(contadorInodo, InodoAux)
	contadorInodo++

	for i := 0; i < 4; i++ {

		if InodoAux.ApuntadoresBloques[i] > 0 {

			cadenaArchivo += fmt.Sprintf(`Inodo%v:%v->Bloque%v
			
				`, NoInodo, i, contadorBloque)

			//Con el valor del apuntador leemos un struct Bloque
			BloqueAux := estructuras.BloqueDatos{}
			file.Seek(int64(InodoAux.ApuntadoresBloques[i]+int32(1)), 0)
			SizeBloque := int(unsafe.Sizeof(BloqueAux))
			BloqueData := leerBytes(file, int(SizeBloque))
			buffer := bytes.NewBuffer(BloqueData)
			err := binary.Read(buffer, binary.BigEndian, &BloqueAux)
			if err != nil {
				fmt.Println(err)
				return

			}

			cadenaArchivo += GenerarBloque(contadorBloque, &BloqueAux)
			contadorBloque++

		}

	}

	if InodoAux.ApuntadorIndirecto > 0 {

		cadenaArchivo += fmt.Sprintf(`Inodo%v:4->Inodo%v
			
				`, NoInodo, contadorInodo)

		//Con el valor del apuntador leemos un struct Inodo
		InodoExt := estructuras.Inodo{}
		file.Seek(int64(InodoAux.ApuntadorIndirecto+int32(1)), 0)
		SizeInodo := int(unsafe.Sizeof(InodoExt))
		ExtData := leerBytes(file, int(SizeInodo))
		buffer := bytes.NewBuffer(ExtData)
		err := binary.Read(buffer, binary.BigEndian, &InodoExt)
		if err != nil {
			fmt.Println(err)
			return

		}

		EscribirInodoRecursivo(file, &InodoExt, contadorInodo)
	}

}

/////////////////////////////////////////////////////////////////////////////////

//ReporteTreeComplete crea el reporte del sistema completo Generarbloque
func ReporteTreeComplete3(path string, ruta string, id string) {

	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}

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

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDroot := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDroot))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				RootData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(RootData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDroot)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				EscribirTreeComplete3(fileMBR, &AVDroot, extension, path)

				color.Print("@{w}El reporte@{w} inode @{w}fue creado con éxito\n")

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

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDroot := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDroot))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				RootData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(RootData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDroot)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				EscribirTreeComplete3(fileMBR, &AVDroot, extension, path)

				color.Print("@{w}El reporte@{w} inode @{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte inode solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}

}

//EscribirTreeComplete genera el reporte TreeComplete al recibir la AVD de la raiz
func EscribirTreeComplete3(MBRfile *os.File, AVDroot *estructuras.AVD, extension string, path string) {

	file, err := os.OpenFile("codigo45.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	// Change permissions Linux.
	err = os.Chmod("codigo45.dot", 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	file.Truncate(0)
	file.Seek(0, 0)

	w := bufio.NewWriter(file)

	fmt.Fprint(w, `digraph Tree {
		node [shape=plaintext];
		`)

	///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

	contadorAVD = 0
	contadorDD = 0
	contadorInodo = 0
	contadorBloque = 0
	cadenaArchivo = ""

	EscribirAVDRecursivo3(MBRfile, AVDroot, contadorAVD)

	fmt.Fprint(w, cadenaArchivo)

	///////////////////////////////////////////////////////////////////////////////////////

	fmt.Fprint(w, `}`)

	w.Flush()

	file.Close()

	extT := "-T"

	switch strings.ToLower(extension) {
	case ".png":
		extT += "png"
	case ".pdf":
		extT += "pdf"
	case ".jpg":
		extT += "jpg"
	default:

	}

	carpeta := filepath.Dir(path)
	SVGpath := carpeta + "/inode.svg"

	if runtime.GOOS == "windows" {
		cmd := exec.Command("dot", extT, "-o", path, "codigo45.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo45.dot")
		cmd2.Run()
	} else {
		cmd := exec.Command("dot", extT, "-o", path, "codigo45.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo45.dot")
		cmd2.Run()
	}

}

//EscribirAVDRecursivo recorre un AVD
func EscribirAVDRecursivo3(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	//cadenaArchivo += GenerarAVD3(contadorAVD, AVDAux)
	contadorAVD++

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

			//	cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v

			//		`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo3(file, &AVDHijo, contadorAVD)

		}
	}

	//cadenaArchivo += fmt.Sprintf(`AVD%v:6->DD%v

	//		`, NoAVD, contadorDD)

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

	n := bytes.Index(AVDAux.NombreDir[:], []byte{0})
	if n == -1 {
		n = len(AVDAux.NombreDir)
	}
	carpeta := string(AVDAux.NombreDir[:n])
	EscribirDDRecursivo3(file, &DDAux, contadorDD, carpeta)

	if AVDAux.ApuntadorAVD > 0 {

		//	cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

		//		`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo3(file, &AVDExt, contadorAVD)
	}

}

//EscribirAVDExtRecursivo recorre la extensión del AVD
func EscribirAVDExtRecursivo3(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	//	cadenaArchivo += GenerarAVD3(contadorAVD, AVDAux)
	contadorAVD++

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

			//cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v

			//	`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo3(file, &AVDHijo, contadorAVD)

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

		//cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

		//		`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo3(file, &AVDExt, contadorAVD)
	}

}

//EscribirDDRecursivo recorre el detalle de directorio
func EscribirDDRecursivo3(file *os.File, DDaux *estructuras.DD, NoDD int, carpeta string) {

	//cadenaArchivo += GenerarDD3(NoDD, DDaux, carpeta)
	contadorDD++

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

			//	cadenaArchivo += fmt.Sprintf(`DD%v:%v->Inodo%v

			//	`, NoDD, i, contadorInodo)

			EscribirInodoRecursivo3(file, &InodoAux, contadorInodo)

		}
	}

	if DDaux.ApuntadorDD > 0 {

		//	cadenaArchivo += fmt.Sprintf(`DD%v:5->DD%v

		//	`, NoDD, contadorDD)

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

		EscribirDDRecursivo3(file, &DDExt, contadorDD, carpeta)
	}

}

//////////////////////////////////////////////////////////////////////////////
//ReporteTreeDirectorio genera el reporte de un directorio
func ReporteTreeDirectorio(path string, ruta string, id string) {
	color.Println(path + " r" + ruta + " " + id)
	if ruta == "" {
		ruta = "/"
	}
	if ruta != "" {

		if strings.HasPrefix(ruta, "/") {

			if ruta != "/" { // si no es root quita slash al final
				if last := len(ruta) - 1; last >= 0 && ruta[last] == '/' {
					ruta = ruta[:last]
				}
			}

			extension := filepath.Ext(path)

			if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

				NameAux, PathAux := GetDatosPart(id)

				if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

					//LEER Y RECORRER EL MBR
					fileMBR, err2 := os.Open(PathAux)
					if err2 != nil { //validar que no sea nulo.
						panic(err2)
					}

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

						//En esta parte debemos ir recorriendo la ruta que recibimos como parámetro

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
						carpetas := strings.Split(ruta, "/")
						i := 1
						PathCorrecto := true

						for i < len(carpetas)-1 {

							if TieneSub, ApuntadorSiguiente := ExisteSub(carpetas[i], int(ApuntadorAVD), PathAux); TieneSub {

								//Si entramos a esta parte, significa que el padre si contiene al hijo (subdirectorio)
								//El hijo sería otro padre en el path o directamente será el padre de la carpeta que queremos crear
								//Por lo tanto leeremos otro AVD con el resultado de "APuntadorSiguiente" y seguiremos.

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
								copy(NombreAnterior[:], AVDAux.NombreDir[:])
								i++
								PathCorrecto = true

							} else {
								PathCorrecto = false
								break
							}
						}

						if PathCorrecto {

							if ruta != "/" {
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

									EscribirTreeDirectorio(fileMBR, &AVDAux, extension, path)

									color.Print("@{w}El reporte@{w} inode @{w}fue creado con éxito\n")

								} else {
									color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
								}
							} else {

								EscribirTreeDirectorio(fileMBR, &AVDAux, extension, path)

								color.Print("@{w}El reporte@{w} inode @{w}fue creado con éxito\n")
							}

						} else {
							color.Println("@{r} Error, una o más carpetas padre no existen.")
						}

						/////////////////////////////////////////////////////////////////////////////

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

						//En esta parte debemos ir recorriendo la ruta que recibimos como parámetro

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
						carpetas := strings.Split(ruta, "/")
						i := 1
						PathCorrecto := true

						for i < len(carpetas)-1 {

							if TieneSub, ApuntadorSiguiente := ExisteSub(carpetas[i], int(ApuntadorAVD), PathAux); TieneSub {

								//Si entramos a esta parte, significa que el padre si contiene al hijo (subdirectorio)
								//El hijo sería otro padre en el path o directamente será el padre de la carpeta que queremos crear
								//Por lo tanto leeremos otro AVD con el resultado de "APuntadorSiguiente" y seguiremos.

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
								copy(NombreAnterior[:], AVDAux.NombreDir[:])
								i++
								PathCorrecto = true

							} else {
								PathCorrecto = false
								break
							}
						}

						if PathCorrecto {

							if ruta != "/" {
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

									EscribirTreeDirectorio(fileMBR, &AVDAux, extension, path)

									color.Print("@{w}El reporte@{w} inode @{w}fue creado con éxito\n")

								} else {
									color.Printf("@{r}La carpeta @{w}%v @{r}no existe.\n", carpetas[len(carpetas)-1])
								}
							} else {

								EscribirTreeDirectorio(fileMBR, &AVDAux, extension, path)

								color.Print("@{w}El reporte@{w} inode @{w}fue creado con éxito\n")
							}

						} else {
							color.Println("@{r} Error, una o más carpetas padre no existen.")
						}

						/////////////////////////////////////////////////////////////////////////////

					} else {
						color.Println("@{r} La partición indicada no ha sido formateada.")
					}

					fileMBR.Close()

				}

			} else {
				color.Println("@{r}El reporte inode solo puede generar archivos con extensión .png, .jpg ó .pdf.")
			}

		} else {
			color.Println("@{r}Ruta incorrecta, debe iniciar con @{w}/")
		}

	} else {
		color.Println("@{r}Faltan parámetros obligatorios para el reporte inode.")
	}

}

//EscribirTreeDirectorio genera el reporte TreeComplete al recibir la AVD de la raiz
func EscribirTreeDirectorio(MBRfile *os.File, AVDroot *estructuras.AVD, extension string, path string) {

	file, err := os.OpenFile("codigo2.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	// Change permissions Linux.
	err = os.Chmod("codigo2.dot", 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	file.Truncate(0)
	file.Seek(0, 0)

	w := bufio.NewWriter(file)

	fmt.Fprint(w, `digraph Tree {
		node [shape=plaintext];
		`)

	///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

	contadorAVD = 0
	contadorDD = 0
	contadorInodo = 0
	contadorBloque = 0
	cadenaArchivo = ""

	EscribirAVDRecursivo2(MBRfile, AVDroot, contadorAVD)

	fmt.Fprint(w, cadenaArchivo)

	///////////////////////////////////////////////////////////////////////////////////////

	fmt.Fprint(w, `}`)

	w.Flush()

	file.Close()

	extT := "-T"

	switch strings.ToLower(extension) {
	case ".png":
		extT += "png"
	case ".pdf":
		extT += "pdf"
	case ".jpg":
		extT += "jpg"
	default:

	}

	carpeta := filepath.Dir(path)
	SVGpath := carpeta + "/treedirectorio.svg"

	if runtime.GOOS == "windows" {
		cmd := exec.Command("dot", extT, "-o", path, "codigo2.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo2.dot")
		cmd2.Run()
	} else {
		cmd := exec.Command("dot", extT, "-o", path, "codigo2.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo2.dot")
		cmd2.Run()
	}

}

//EscribirAVDRecursivo2 recorre un AVD
func EscribirAVDRecursivo2(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	//cadenaArchivo += GenerarAVD(contadorAVD, AVDAux)
	contadorAVD++

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

			//	cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v

			//	`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo2(file, &AVDHijo, contadorAVD)

		}
	}

	cadenaArchivo += fmt.Sprintf(`AVD%v:6->DD%v

			`, NoAVD, contadorDD)

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

	n := bytes.Index(AVDAux.NombreDir[:], []byte{0})
	if n == -1 {
		n = len(AVDAux.NombreDir)
	}
	carpeta := string(AVDAux.NombreDir[:n])
	EscribirDDRecursivo2(file, &DDAux, contadorDD, carpeta)

	if AVDAux.ApuntadorAVD > 0 {

		cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

			`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo2(file, &AVDExt, contadorAVD)
	}

}

//EscribirAVDExtRecursivo2 recorre la extensión del AVD
func EscribirAVDExtRecursivo2(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	//cadenaArchivo += GenerarAVD(contadorAVD, AVDAux)
	contadorAVD++

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

			//cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v

			//	`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo2(file, &AVDHijo, contadorAVD)

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

		//	cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

		//	`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo2(file, &AVDExt, contadorAVD)
	}

}

//EscribirDDRecursivo2 recorre el detalle de directorio
func EscribirDDRecursivo2(file *os.File, DDaux *estructuras.DD, NoDD int, carpeta string) {

	cadenaArchivo += GenerarDD(NoDD, DDaux, carpeta)
	contadorDD++

	if DDaux.ApuntadorDD > 0 {

		cadenaArchivo += fmt.Sprintf(`DD%v:5->DD%v
	
		`, NoDD, contadorDD)

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

		EscribirDDRecursivo2(file, &DDExt, contadorDD, carpeta)
	}

}

//ReporteTreeFile genera el reporte de un archivo
func ReporteTreeFile(path string, ruta string, id string) {
	color.Println(path + " r" + ruta + " " + id)
	if ruta == "" {
		ruta = "/"
	}
	if ruta != "" {

		if strings.HasPrefix(ruta, "/") {

			extension := filepath.Ext(path)

			if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

				NameAux, PathAux := GetDatosPart(id)

				if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

					//LEER Y RECORRER EL MBR
					fileMBR, err2 := os.Open(PathAux)
					if err2 != nil { //validar que no sea nulo.
						panic(err2)
					}

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

						if ReporteExitoso := EscribirTreeFile(fileMBR, &SB1, extension, path, PathAux, ruta); ReporteExitoso {
							color.Print("@{w}El reporte@{w} TREE_FILE @{w}fue creado con éxito\n")
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

						if ReporteExitoso := EscribirTreeFile(fileMBR, &SB1, extension, path, PathAux, ruta); ReporteExitoso {
							color.Print("@{w}El reporte@{w} TREE_FILE @{w}fue creado con éxito\n")
						}

					} else {
						color.Println("@{r} La partición indicada no ha sido formateada.")
					}

					fileMBR.Close()

				}

			} else {
				color.Println("@{r}El reporte TREE_FILE solo puede generar archivos con extensión .png, .jpg ó .pdf.")
			}

		} else {
			color.Println("@{r}Ruta incorrecta, debe iniciar con @{w}/")
		}

	} else {
		color.Println("@{r}Faltan parámetros obligatorios para el reporte TREE_FILE.")
	}
}

//EscribirTreeFile escribe el archivo para tree_file
func EscribirTreeFile(fileMBR *os.File, SB1 *estructuras.Superblock, extension string, path string, PathAux string, ruta string) bool {

	ExisteArchivo := true

	file, err := os.OpenFile("codigo8.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
	if err != nil {
		fmt.Println(err)
		file.Close()
		return false
	}

	// Change permissions Linux.
	err = os.Chmod("codigo8.dot", 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
		return false
	}

	file.Truncate(0)
	file.Seek(0, 0)

	w := bufio.NewWriter(file)

	fmt.Fprint(w, `digraph Tree {
		node [shape=plaintext];
		`)

	//////// SETEAMOS VARIABLES A CERO ////////////////////////////

	contadorAVD = 0
	contadorDD = 0
	contadorInodo = 0
	contadorBloque = 0
	cadenaArchivo = ""

	///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

	//En esta parte debemos ir recorriendo la ruta que recibimos como parámetro

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
		return false
	}

	cadenaArchivo += GenerarAVD(contadorAVD, &AVDAux)
	contadorAVD++

	//Vamos a comparar Padres e Hijos
	carpetas := strings.Split(ruta, "/")
	i := 1

	PathCorrecto := true
	NoAVD := 0
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
						return false
					}
					//Comparamos el nombre del AVD leido con el nombre del directorio que queremos verificar si existe
					//si existe el directorio retornamos true y el byte donde está dicho AVD
					var chars [20]byte
					copy(chars[:], carpetas[i])

					if string(AVDHijo.NombreDir[:]) == string(chars[:]) {

						cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v
			
						`, NoAVD, x, contadorAVD)

						cadenaArchivo += GenerarAVD(contadorAVD, &AVDHijo)
						contadorAVD++
						NoAVD++

						ApuntadorAVD = int32(AVDAux.ApuntadorSubs[x])
						fileMBR.Seek(int64(ApuntadorAVD+1), 0)
						AnteriorData = leerBytes(fileMBR, int(SizeAVD))
						buffer2 = bytes.NewBuffer(AnteriorData)
						err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
						if err != nil {
							fileMBR.Close()
							fmt.Println(err)
							return false
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

				cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

				`, NoAVD, contadorAVD)

				//Leemos el AVD (que se considera contiguo)
				fileMBR.Seek(int64(AVDAux.ApuntadorAVD+int32(1)), 0)
				AnteriorData = leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(AnteriorData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return false
				}

				cadenaArchivo += GenerarAVD(contadorAVD, &AVDAux)
				contadorAVD++
				NoAVD++

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

		NoDD := 0

		cadenaArchivo += fmt.Sprintf(`AVD%v:6->DD%v

			`, NoAVD, contadorDD)

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
			return false
		}

		n := bytes.Index(AVDAux.NombreDir[:], []byte{0})
		if n == -1 {
			n = len(AVDAux.NombreDir)
		}
		folder := string(AVDAux.NombreDir[:n])

		cadenaArchivo += GenerarDD(NoDD, &DDAux, folder)
		contadorDD++

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
							return false
						}

						cadenaArchivo += fmt.Sprintf(`DD%v:%v->Inodo%v
			
						`, NoDD, i, contadorInodo)

						EscribirInodoRecursivo(fileMBR, &InodoAux, contadorInodo)

						Continuar = false
						break

					}

				}

			}

			if Continuar == false {
				continue
			}

			//Si el archivo no está en el arreglo de archivos
			//verificamos si el DD actual apunta hacia otro DD

			if DDAux.ApuntadorDD > 0 {

				cadenaArchivo += fmt.Sprintf(`DD%v:5->DD%v
	
				`, NoDD, contadorDD)

				//Leemos el DD (que se considera contiguo)
				fileMBR.Seek(int64(DDAux.ApuntadorDD+int32(1)), 0)
				DDData = leerBytes(fileMBR, int(SizeDD))
				bufferDD = bytes.NewBuffer(DDData)
				err = binary.Read(bufferDD, binary.BigEndian, &DDAux)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return false
				}

				cadenaArchivo += GenerarDD(contadorDD, &DDAux, folder)
				contadorDD++
				NoDD++

			} else {
				//Si ya no apunta a otro DD y llegamos a esta parte, cancelamos el ciclo FOR
				Continuar = false
				ExisteArchivo = false
				color.Println("@{r} El archivo no existe.")
				break
			}
		}

	} else {
		color.Println("@{r} Error, una o más carpetas padre no existen.")
		ExisteArchivo = false
	}

	///////////////////////////////////////////////////////////////////////////////////////

	fmt.Fprint(w, cadenaArchivo)

	fmt.Fprint(w, `}`)

	w.Flush()

	file.Close()

	if ExisteArchivo {
		extT := "-T"

		switch strings.ToLower(extension) {
		case ".png":
			extT += "png"
		case ".pdf":
			extT += "pdf"
		case ".jpg":
			extT += "jpg"
		default:

		}

		carpeta := filepath.Dir(path)
		SVGpath := carpeta + "/treefile.svg"

		if runtime.GOOS == "windows" {
			cmd := exec.Command("dot", extT, "-o", path, "codigo8.dot")
			cmd.Run()
			cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo8.dot")
			cmd2.Run()
		} else {
			cmd := exec.Command("dot", extT, "-o", path, "codigo8.dot")
			cmd.Run()
			cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo8.dot")
			cmd2.Run()
		}

		return true
	}

	return false

}

//ReporteBitmapAVD crea el reporte del Bitmap AVD
func ReporteBitmapAVD(path string, ruta string, id string) {
	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".txt" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}
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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitmapAVDS+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalAVDS))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_ARBDIR@{w}fue creado con éxito\n")

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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitmapAVDS+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalAVDS))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_ARBDIR@{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte BM_ARBDIR solo puede generar archivos con extensión .txt.")
	}
}

//ReporteBitmapDD crea el reporte del Bitmap AVD
func ReporteBitmapDD(path string, ruta string, id string) {
	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".txt" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}
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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitMapDDS+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalDDS))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_DETDIR@{w}fue creado con éxito\n")

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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitMapDDS+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalDDS))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_DETDIR@{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte BM_DETDIR solo puede generar archivos con extensión .txt.")
	}
}

//ReporteBitmapInode crea el reporte del Bitmap AVD
func ReporteBitmapInode(path string, ruta string, id string) {
	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".txt" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}
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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitmapInodos+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalInodos))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_INODE@{w}fue creado con éxito\n")

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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitmapInodos+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalInodos))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_INODE@{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte BM_INODE solo puede generar archivos con extensión .txt.")
	}
}

//ReporteBitmapBloque crea el reporte del Bitmap AVD
func ReporteBitmapBloque(path string, ruta string, id string) {
	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".txt" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}
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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitmapBloques+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalBloques))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_BLOCK@{w}fue creado con éxito\n")

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

				file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}
				// Change permissions Linux.
				err = os.Chmod(path, 0666)
				if err != nil {
					fmt.Println(err)
					file.Close()
					return
				}

				file.Truncate(0)
				file.Seek(0, 0)
				w := bufio.NewWriter(file)

				fileMBR.Seek(int64(SB1.InicioBitmapBloques+1), 0)
				BitmapData := leerBytes(fileMBR, int(SB1.TotalBloques))
				contador := 0
				for _, b := range BitmapData {
					if b == 1 {
						fmt.Fprint(w, "1	")
					} else {
						fmt.Fprint(w, "0	")
					}
					contador++
					if contador == 20 {
						fmt.Fprint(w, "\n")
						contador = 0
					}
				}

				w.Flush()
				file.Close()

				color.Print("@{w}El reporte@{w} BM_BLOCK@{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte BM_BLOCK solo puede generar archivos con extensión .txt.")
	}
}
func EscribirInodoRecursivo3(file *os.File, InodoAux *estructuras.Inodo, NoInodo int) {

	cadenaArchivo += GenerarInodo(contadorInodo, InodoAux)
	contadorInodo++

	for i := 0; i < 4; i++ {

		if InodoAux.ApuntadoresBloques[i] > 0 {

			//Con el valor del apuntador leemos un struct Bloque
			BloqueAux := estructuras.BloqueDatos{}
			file.Seek(int64(InodoAux.ApuntadoresBloques[i]+int32(1)), 0)
			SizeBloque := int(unsafe.Sizeof(BloqueAux))
			BloqueData := leerBytes(file, int(SizeBloque))
			buffer := bytes.NewBuffer(BloqueData)
			err := binary.Read(buffer, binary.BigEndian, &BloqueAux)
			if err != nil {
				fmt.Println(err)
				return

			}

			//cadenaArchivo += GenerarBloque(contadorBloque, &BloqueAux)
			contadorBloque++

		}

	}

	if InodoAux.ApuntadorIndirecto > 0 {

		cadenaArchivo += fmt.Sprintf(`Inodo%v:4->Inodo%v
			
				`, NoInodo, contadorInodo)

		//Con el valor del apuntador leemos un struct Inodo
		InodoExt := estructuras.Inodo{}
		file.Seek(int64(InodoAux.ApuntadorIndirecto+int32(1)), 0)
		SizeInodo := int(unsafe.Sizeof(InodoExt))
		ExtData := leerBytes(file, int(SizeInodo))
		buffer := bytes.NewBuffer(ExtData)
		err := binary.Read(buffer, binary.BigEndian, &InodoExt)
		if err != nil {
			fmt.Println(err)
			return

		}

		EscribirInodoRecursivo3(file, &InodoExt, contadorInodo)
	}

}
func EscribirInodoRecursivo4(file *os.File, InodoAux *estructuras.Inodo, NoInodo int) {

	//cadenaArchivo += GenerarInodo(contadorInodo, InodoAux)
	contadorInodo++

	for i := 0; i < 4; i++ {

		if InodoAux.ApuntadoresBloques[i] > 0 {

			//Con el valor del apuntador leemos un struct Bloque
			BloqueAux := estructuras.BloqueDatos{}
			file.Seek(int64(InodoAux.ApuntadoresBloques[i]+int32(1)), 0)
			SizeBloque := int(unsafe.Sizeof(BloqueAux))
			BloqueData := leerBytes(file, int(SizeBloque))
			buffer := bytes.NewBuffer(BloqueData)
			err := binary.Read(buffer, binary.BigEndian, &BloqueAux)
			if err != nil {
				fmt.Println(err)
				return

			}

			cadenaArchivo += GenerarBloque(contadorBloque, &BloqueAux)
			contadorBloque++

		}

	}

	if InodoAux.ApuntadorIndirecto > 0 {

		//	cadenaArchivo += fmt.Sprintf(`Inodo%v:4->Inodo%v

		//			`, NoInodo, contadorInodo)

		//Con el valor del apuntador leemos un struct Inodo
		InodoExt := estructuras.Inodo{}
		file.Seek(int64(InodoAux.ApuntadorIndirecto+int32(1)), 0)
		SizeInodo := int(unsafe.Sizeof(InodoExt))
		ExtData := leerBytes(file, int(SizeInodo))
		buffer := bytes.NewBuffer(ExtData)
		err := binary.Read(buffer, binary.BigEndian, &InodoExt)
		if err != nil {
			fmt.Println(err)
			return

		}

		EscribirInodoRecursivo4(file, &InodoExt, contadorInodo)
	}

}

//ReporteTreeComplete crea el reporte del sistema completo Generarbloque
func ReporteTreeComplete4(path string, ruta string, id string) {

	extension := filepath.Ext(path)

	if strings.ToLower(extension) == ".pdf" || strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".png" {

		NameAux, PathAux := GetDatosPart(id)

		if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe {

			//LEER Y RECORRER EL MBR
			fileMBR, err2 := os.Open(PathAux)
			if err2 != nil { //validar que no sea nulo.
				panic(err2)
			}

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

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDroot := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDroot))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				RootData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(RootData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDroot)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				EscribirTreeComplete4(fileMBR, &AVDroot, extension, path)

				color.Print("@{w}El reporte@{w} block @{w}fue creado con éxito\n")

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

				//NOS POSICIONAMOS DONDE EMPIEZA EL STRUCT DE LA CARPETA ROOT (primer struct AVD)
				ApuntadorAVD := SB1.InicioAVDS
				//CREAMOS UN STRUCT TEMPORAL
				AVDroot := estructuras.AVD{}
				SizeAVD := int(unsafe.Sizeof(AVDroot))
				fileMBR.Seek(int64(ApuntadorAVD+1), 0)
				RootData := leerBytes(fileMBR, int(SizeAVD))
				buffer2 := bytes.NewBuffer(RootData)
				err = binary.Read(buffer2, binary.BigEndian, &AVDroot)
				if err != nil {
					fileMBR.Close()
					fmt.Println(err)
					return
				}

				EscribirTreeComplete4(fileMBR, &AVDroot, extension, path)

				color.Print("@{w}El reporte@{w} block @{w}fue creado con éxito\n")

			} else {
				color.Println("@{r} La partición indicada no ha sido formateada.")
			}

			fileMBR.Close()
		}

	} else {
		color.Println("@{r}El reporte block solo puede generar archivos con extensión .png, .jpg ó .pdf.")
	}

}

//EscribirTreeComplete genera el reporte TreeComplete al recibir la AVD de la raiz
func EscribirTreeComplete4(MBRfile *os.File, AVDroot *estructuras.AVD, extension string, path string) {

	file, err := os.OpenFile("codigo46.dot", os.O_CREATE|os.O_RDWR, 0666) //Crea un nuevo archivo
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	// Change permissions Linux.
	err = os.Chmod("codigo46.dot", 0666)
	if err != nil {
		fmt.Println(err)
		file.Close()
		return
	}

	file.Truncate(0)
	file.Seek(0, 0)

	w := bufio.NewWriter(file)

	fmt.Fprint(w, `digraph Tree {
		node [shape=plaintext];
		`)

	///////// AQUI COMENZAMOS A RECORRER TODO EL SISTEMA LWH /////////////////////////////////

	contadorAVD = 0
	contadorDD = 0
	contadorInodo = 0
	contadorBloque = 0
	cadenaArchivo = ""

	EscribirAVDRecursivo4(MBRfile, AVDroot, contadorAVD)

	fmt.Fprint(w, cadenaArchivo)

	///////////////////////////////////////////////////////////////////////////////////////

	fmt.Fprint(w, `}`)

	w.Flush()

	file.Close()

	extT := "-T"

	switch strings.ToLower(extension) {
	case ".png":
		extT += "png"
	case ".pdf":
		extT += "pdf"
	case ".jpg":
		extT += "jpg"
	default:

	}

	carpeta := filepath.Dir(path)
	SVGpath := carpeta + "/block.svg"

	if runtime.GOOS == "windows" {
		cmd := exec.Command("dot", extT, "-o", path, "codigo46.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo46.dot")
		cmd2.Run()
	} else {
		cmd := exec.Command("dot", extT, "-o", path, "codigo46.dot")
		cmd.Run()
		cmd2 := exec.Command("dot", "-Tsvg", "-o", SVGpath, "codigo46.dot")
		cmd2.Run()
	}

}

//EscribirAVDRecursivo recorre un AVD
func EscribirAVDRecursivo4(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	//cadenaArchivo += GenerarAVD3(contadorAVD, AVDAux)
	contadorAVD++

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

			//	cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v

			//		`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo4(file, &AVDHijo, contadorAVD)

		}
	}

	//cadenaArchivo += fmt.Sprintf(`AVD%v:6->DD%v

	//		`, NoAVD, contadorDD)

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

	n := bytes.Index(AVDAux.NombreDir[:], []byte{0})
	if n == -1 {
		n = len(AVDAux.NombreDir)
	}
	carpeta := string(AVDAux.NombreDir[:n])
	EscribirDDRecursivo4(file, &DDAux, contadorDD, carpeta)

	if AVDAux.ApuntadorAVD > 0 {

		//	cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

		//		`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo4(file, &AVDExt, contadorAVD)
	}

}

//EscribirAVDExtRecursivo recorre la extensión del AVD
func EscribirAVDExtRecursivo4(file *os.File, AVDAux *estructuras.AVD, NoAVD int) {

	//	cadenaArchivo += GenerarAVD3(contadorAVD, AVDAux)
	contadorAVD++

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

			//cadenaArchivo += fmt.Sprintf(`AVD%v:%v->AVD%v

			//	`, NoAVD, i, contadorAVD)

			EscribirAVDRecursivo4(file, &AVDHijo, contadorAVD)

		}
	}

	if AVDAux.ApuntadorAVD > 0 {

		//cadenaArchivo += fmt.Sprintf(`AVD%v:7->AVD%v

		//		`, NoAVD, contadorAVD)

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

		EscribirAVDExtRecursivo4(file, &AVDExt, contadorAVD)
	}

}

//EscribirDDRecursivo recorre el detalle de directorio
func EscribirDDRecursivo4(file *os.File, DDaux *estructuras.DD, NoDD int, carpeta string) {

	//cadenaArchivo += GenerarDD3(NoDD, DDaux, carpeta)
	contadorDD++

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

			//	cadenaArchivo += fmt.Sprintf(`DD%v:%v->Inodo%v

			//	`, NoDD, i, contadorInodo)

			EscribirInodoRecursivo4(file, &InodoAux, contadorInodo)

		}
	}

	if DDaux.ApuntadorDD > 0 {

		//	cadenaArchivo += fmt.Sprintf(`DD%v:5->DD%v

		//	`, NoDD, contadorDD)

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

		EscribirDDRecursivo4(file, &DDExt, contadorDD, carpeta)
	}

}
