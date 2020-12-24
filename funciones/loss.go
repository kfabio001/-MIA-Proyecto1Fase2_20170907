package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarLoss function
func EjecutarLoss(id string) {

	if sesionRoot {

		if id != "" {

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

						//Antes de borrar todos los datos, debemos hacer un copia temporal del bloque bitacoras y del Backup del super bloque

						fileMBR.Seek(int64(SB1.InicioBitacora+1), 0)
						CantidadBytes := int(SB1.TotalBitacoras*SB1.SizeBitacora) + int(unsafe.Sizeof(SB1))
						//Toda la info de las bitacoras y el backUp se almacena en BackupData
						BackUpData := leerBytes(fileMBR, CantidadBytes)

						//Toca setear el superbloque auxiliar

						SB1.FreeAVDS = SB1.TotalAVDS - int32(1)
						SB1.FreeDDS = SB1.TotalDDS - int32(1)
						SB1.FreeInodos = SB1.TotalInodos - int32(1)                       //Restamos 1 bloque porque al escribir users.txt ocupamos 1 bloque
						SB1.FreeBloques = SB1.TotalBloques - int32(2)                     //Restamos 2 bloques porque al escribir users.txt ocupamos 2 bloques
						SB1.FirstFreeAVD = SB1.InicioAVDS + SB1.SizeAVD                   //Le sumamos un sizeAVD porque vamos a crear la carpeta "/"
						SB1.FirstFreeDD = SB1.InicioDDS + SB1.SizeDD                      //Le sumamos un sizeDD porque vamos a crear el DD de la carpeta "/"
						SB1.FirstFreeInodo = SB1.InicioInodos + SB1.SizeInodo             //Le sumamos un sizeInodo porque vamos a crear el inodo del archivo users.txt
						SB1.FirstFreeBloque = SB1.InicioBloques + int32(2*SB1.SizeBloque) //Le sumamos dos sizeBloque porque, vamos a crear un usuario y un grupo default, lo cual abarca 32 caracteres.

						//Rellenamos toda la particion de 0's
						fileMBR.Seek(int64(SB1.PartStart+1), 0)
						data := make([]byte, SB1.PartSize)
						fileMBR.Write(data)

						//Escribiendo el Superbloque reseteado
						fileMBR.Seek(int64(SB1.PartStart+1), 0)
						sb1 := &SB1
						var binario1 bytes.Buffer
						binary.Write(&binario1, binary.BigEndian, sb1)
						escribirBytes(fileMBR, binario1.Bytes())

						//Escribiendo el Bloque bitacoras original y el bakcup del super bloque original
						fileMBR.Seek(int64(SB1.InicioBitacora+1), 0)
						fileMBR.Write(BackUpData)

						//Volviendo a crear carpeta "/" y users.txt

						//Bitmap de AVD (la primera posición es la carpeta "/")
						//Escribiendo un 1 en la primera posicion del bitmap
						fileMBR.Seek(int64(SB1.InicioBitmapAVDS+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)
						//Seteando valores de la carpeta root "/
						//Creamos nueva estructura AVD, la cual será escrita en su posición correspondiente
						AVDaux := estructuras.AVD{}
						t := time.Now()
						var charsDate [20]byte
						cadena := t.Format("2006-01-02 15:04:05")
						copy(charsDate[:], cadena)
						//Seteando fecha de creacion
						copy(AVDaux.FechaCreacion[:], charsDate[:])
						var ArrayNombre [20]byte
						nombreDir := "/"
						copy(ArrayNombre[:], nombreDir)
						//Seteando nombre del directorio "/"
						copy(AVDaux.NombreDir[:], ArrayNombre[:])
						//La primera estructura AVD apuntará al primer Detalle de Directorio
						//Seteando el apuntador de su DD, en este caso es InicioDDS
						//al ser el primer DD que se usar
						AVDaux.ApuntadorDD = SB1.InicioDDS
						AVDaux.ApuntadorSubs[0] = 0
						AVDaux.ApuntadorSubs[1] = 0
						AVDaux.ApuntadorSubs[2] = 0
						AVDaux.ApuntadorSubs[3] = 0
						AVDaux.ApuntadorSubs[4] = 0
						AVDaux.ApuntadorSubs[5] = 0
						AVDaux.PermisoU = 7
						AVDaux.PermisoG = 7
						AVDaux.PermisoO = 7
						nombrePropietario := "root"
						var ArrayPropietario [20]byte
						copy(ArrayPropietario[:], nombrePropietario)
						//Seteando nombre del propietario, en este caso la raiz pertenece al id "root"
						copy(AVDaux.Proper[:], ArrayPropietario[:])
						//APuntadorAVD y los 6 apuntadores a subdirectorios no se setean en este momento
						//se hará conforme se vayan creando subdirectorios :)

						//Seteando nombre del grupo, en este caso pertenece al grupo root
						var ArrayGrupo [20]byte
						nombreGrupo := "root"
						copy(ArrayGrupo[:], nombreGrupo)
						copy(AVDaux.Grupo[:], ArrayGrupo[:])
						//Ahora toca escribir el struct AVD en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioAVDS+1), 0)
						avdp := &AVDaux
						var binario3 bytes.Buffer
						binary.Write(&binario3, binary.BigEndian, avdp)
						escribirBytes(fileMBR, binario3.Bytes())

						//Bitmap de DD (la primera posición es el DD de la carpeta "/")
						//Escribiendo un 1 en la primera posicion del bitmap
						fileMBR.Seek(int64(SB1.InicioBitMapDDS+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)
						//A continuación debemos crear el archivo users.txt
						//Creamos una estructura DD para la carpeta "/"
						DDaux := estructuras.DD{}

						//Seteamos atributos al DD
						nombreArchivo := "users.txt"
						copy(ArrayNombre[:], nombreArchivo)
						//Seteando nombre del archivo, al primer struct del arreglo del DD
						copy(DDaux.DDFiles[0].Name[:], ArrayNombre[:])
						//Seteando los atributos FechaCreacio y FechaModificacion del archivo users.txt
						t = time.Now()
						cadena = t.Format("2006-01-02 15:04:05")
						copy(charsDate[:], cadena)
						copy(DDaux.DDFiles[0].FechaCreacion[:], charsDate[:])
						copy(DDaux.DDFiles[0].FechaModificacion[:], charsDate[:])
						//Seteamos el apuntador al inodo, en este caso es sb.InicioInodos
						//al ser el primer Inodo en usarse
						DDaux.DDFiles[0].ApuntadorInodo = SB1.InicioInodos

						//Ahora toca escribir el struct DD en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioDDS+1), 0)
						ddp := &DDaux
						var binario4 bytes.Buffer
						binary.Write(&binario4, binary.BigEndian, ddp)
						escribirBytes(fileMBR, binario4.Bytes())

						//Bitmap de inodos (la primera posición es el inodo para el archivo users.txt)
						//Escribiendo un 1 en la primera posicion del bitmap
						fileMBR.Seek(int64(SB1.InicioBitmapInodos+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)

						//A continuacion creamos una struct de tipo Inodo
						InodoAux := estructuras.Inodo{}
						//Seteamos atributos al Inodo
						var ArrayPInodo [20]byte
						nombrePropietario = "root"
						copy(ArrayPInodo[:], nombrePropietario)
						copy(InodoAux.Proper[:], ArrayPInodo[:])
						var ArrayGInodo [20]byte
						nombreGrupo = "root"
						copy(ArrayGInodo[:], nombreGrupo)
						copy(InodoAux.Grupo[:], ArrayGInodo[:])
						InodoAux.PermisoU = 7
						InodoAux.PermisoG = 7
						InodoAux.PermisoO = 0
						InodoAux.NumeroInodo = 1
						InodoAux.FileSize = 32
						InodoAux.NumeroBloques = 2
						//Como se creará un usuario y un grupo en users.txt
						//se utilizaran aproximadamente 53 caracteres
						//cada bloque puede almacenar hasta 25 characeres, por lo tanto
						//se necesitarán 2 de los 4 bloques del inodo
						InodoAux.ApuntadoresBloques[0] = SB1.InicioBloques                  // 1,G,root\n1,U,root,root,20 <- caracteres en primer bloque
						InodoAux.ApuntadoresBloques[1] = SB1.InicioBloques + SB1.SizeBloque // 1602625 <- caracteres en segundo bloque

						//Ahora toca escribir el struct Inodo en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioInodos+1), 0)
						inodop := &InodoAux
						var binario5 bytes.Buffer
						binary.Write(&binario5, binary.BigEndian, inodop)
						escribirBytes(fileMBR, binario5.Bytes())

						//Bitmap de inodos (la primeras dos posiciones son para el archivo users.txt)
						//Escribiendo un 1 en las primeras 2 posiciones del bitmap
						fileMBR.Seek(int64(SB1.InicioBitmapBloques+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)
						fileMBR.Seek(int64(SB1.InicioBitmapBloques+2), 0)
						fileMBR.Write(data)

						//A continuación creamos el primer BloqueDatos
						BloqueAux := estructuras.BloqueDatos{}
						contenido := "1,G,root\n1,U,root,root,20"
						copy(BloqueAux.Data[:], contenido)

						//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioBloques+1), 0)
						bloquep := &BloqueAux
						var binario6 bytes.Buffer
						binary.Write(&binario6, binary.BigEndian, bloquep)
						escribirBytes(fileMBR, binario6.Bytes())

						//A continuación creamos el segundo BloqueDatos
						BloqueAux2 := estructuras.BloqueDatos{}
						contenido = "1602625"
						copy(BloqueAux2.Data[:], contenido)
						//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
						fileMBR.Seek(int64((SB1.InicioBloques+SB1.SizeBloque)+1), 0)
						bloque2p := &BloqueAux2
						var binario7 bytes.Buffer
						binary.Write(&binario7, binary.BigEndian, bloque2p)
						escribirBytes(fileMBR, binario7.Bytes())

						color.Println("@{w} COMANDO LOSS realizadO con exito.")

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

						//Antes de borrar todos los datos, debemos hacer un copia temporal del bloque bitacoras y del Backup del super bloque

						fileMBR.Seek(int64(SB1.InicioBitacora+1), 0)
						CantidadBytes := int(SB1.TotalBitacoras*SB1.SizeBitacora) + int(unsafe.Sizeof(SB1))
						//Toda la info de las bitacoras y el backUp se almacena en BackupData
						BackUpData := leerBytes(fileMBR, CantidadBytes)

						//Toca setear el superbloque auxiliar

						SB1.FreeAVDS = SB1.TotalAVDS - int32(1)
						SB1.FreeDDS = SB1.TotalDDS - int32(1)
						SB1.FreeInodos = SB1.TotalInodos - int32(1)                       //Restamos 1 bloque porque al escribir users.txt ocupamos 1 bloque
						SB1.FreeBloques = SB1.TotalBloques - int32(2)                     //Restamos 2 bloques porque al escribir users.txt ocupamos 2 bloques
						SB1.FirstFreeAVD = SB1.InicioAVDS + SB1.SizeAVD                   //Le sumamos un sizeAVD porque vamos a crear la carpeta "/"
						SB1.FirstFreeDD = SB1.InicioDDS + SB1.SizeDD                      //Le sumamos un sizeDD porque vamos a crear el DD de la carpeta "/"
						SB1.FirstFreeInodo = SB1.InicioInodos + SB1.SizeInodo             //Le sumamos un sizeInodo porque vamos a crear el inodo del archivo users.txt
						SB1.FirstFreeBloque = SB1.InicioBloques + int32(2*SB1.SizeBloque) //Le sumamos dos sizeBloque porque, vamos a crear un usuario y un grupo default, lo cual abarca 32 caracteres.

						//Rellenamos toda la particion de 0's
						fileMBR.Seek(int64(SB1.PartStart+1), 0)
						data := make([]byte, SB1.PartSize)
						fileMBR.Write(data)

						//Escribiendo el Superbloque reseteado
						fileMBR.Seek(int64(SB1.PartStart+1), 0)
						sb1 := &SB1
						var binario1 bytes.Buffer
						binary.Write(&binario1, binary.BigEndian, sb1)
						escribirBytes(fileMBR, binario1.Bytes())

						//Escribiendo el Bloque bitacoras original y el bakcup del super bloque original
						fileMBR.Seek(int64(SB1.InicioBitacora+1), 0)
						fileMBR.Write(BackUpData)

						//Volviendo a crear carpeta "/" y users.txt

						//Bitmap de AVD (la primera posición es la carpeta "/")
						//Escribiendo un 1 en la primera posicion del bitmap
						fileMBR.Seek(int64(SB1.InicioBitmapAVDS+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)
						//Seteando valores de la carpeta root "/
						//Creamos nueva estructura AVD, la cual será escrita en su posición correspondiente
						AVDaux := estructuras.AVD{}
						t := time.Now()
						var charsDate [20]byte
						cadena := t.Format("2006-01-02 15:04:05")
						copy(charsDate[:], cadena)
						//Seteando fecha de creacion
						copy(AVDaux.FechaCreacion[:], charsDate[:])
						var ArrayNombre [20]byte
						nombreDir := "/"
						copy(ArrayNombre[:], nombreDir)
						//Seteando nombre del directorio "/"
						copy(AVDaux.NombreDir[:], ArrayNombre[:])
						//La primera estructura AVD apuntará al primer Detalle de Directorio
						//Seteando el apuntador de su DD, en este caso es InicioDDS
						//al ser el primer DD que se usar
						AVDaux.ApuntadorDD = SB1.InicioDDS
						AVDaux.ApuntadorSubs[0] = 0
						AVDaux.ApuntadorSubs[1] = 0
						AVDaux.ApuntadorSubs[2] = 0
						AVDaux.ApuntadorSubs[3] = 0
						AVDaux.ApuntadorSubs[4] = 0
						AVDaux.ApuntadorSubs[5] = 0
						AVDaux.PermisoU = 7
						AVDaux.PermisoG = 7
						AVDaux.PermisoO = 7
						nombrePropietario := "root"
						var ArrayPropietario [20]byte
						copy(ArrayPropietario[:], nombrePropietario)
						//Seteando nombre del propietario, en este caso la raiz pertenece al id "root"
						copy(AVDaux.Proper[:], ArrayPropietario[:])
						//APuntadorAVD y los 6 apuntadores a subdirectorios no se setean en este momento
						//se hará conforme se vayan creando subdirectorios :)

						//Seteando nombre del grupo, en este caso pertenece al grupo root
						var ArrayGrupo [20]byte
						nombreGrupo := "root"
						copy(ArrayGrupo[:], nombreGrupo)
						copy(AVDaux.Grupo[:], ArrayGrupo[:])
						//Ahora toca escribir el struct AVD en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioAVDS+1), 0)
						avdp := &AVDaux
						var binario3 bytes.Buffer
						binary.Write(&binario3, binary.BigEndian, avdp)
						escribirBytes(fileMBR, binario3.Bytes())

						//Bitmap de DD (la primera posición es el DD de la carpeta "/")
						//Escribiendo un 1 en la primera posicion del bitmap
						fileMBR.Seek(int64(SB1.InicioBitMapDDS+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)
						//A continuación debemos crear el archivo users.txt
						//Creamos una estructura DD para la carpeta "/"
						DDaux := estructuras.DD{}

						//Seteamos atributos al DD
						nombreArchivo := "users.txt"
						copy(ArrayNombre[:], nombreArchivo)
						//Seteando nombre del archivo, al primer struct del arreglo del DD
						copy(DDaux.DDFiles[0].Name[:], ArrayNombre[:])
						//Seteando los atributos FechaCreacio y FechaModificacion del archivo users.txt
						t = time.Now()
						cadena = t.Format("2006-01-02 15:04:05")
						copy(charsDate[:], cadena)
						copy(DDaux.DDFiles[0].FechaCreacion[:], charsDate[:])
						copy(DDaux.DDFiles[0].FechaModificacion[:], charsDate[:])
						//Seteamos el apuntador al inodo, en este caso es sb.InicioInodos
						//al ser el primer Inodo en usarse
						DDaux.DDFiles[0].ApuntadorInodo = SB1.InicioInodos

						//Ahora toca escribir el struct DD en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioDDS+1), 0)
						ddp := &DDaux
						var binario4 bytes.Buffer
						binary.Write(&binario4, binary.BigEndian, ddp)
						escribirBytes(fileMBR, binario4.Bytes())

						//Bitmap de inodos (la primera posición es el inodo para el archivo users.txt)
						//Escribiendo un 1 en la primera posicion del bitmap
						fileMBR.Seek(int64(SB1.InicioBitmapInodos+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)

						//A continuacion creamos una struct de tipo Inodo
						InodoAux := estructuras.Inodo{}
						//Seteamos atributos al Inodo
						var ArrayPInodo [20]byte
						nombrePropietario = "root"
						copy(ArrayPInodo[:], nombrePropietario)
						copy(InodoAux.Proper[:], ArrayPInodo[:])
						var ArrayGInodo [20]byte
						nombreGrupo = "root"
						copy(ArrayGInodo[:], nombreGrupo)
						copy(InodoAux.Grupo[:], ArrayGInodo[:])
						InodoAux.PermisoU = 7
						InodoAux.PermisoG = 7
						InodoAux.PermisoO = 0
						InodoAux.NumeroInodo = 1
						InodoAux.FileSize = 32
						InodoAux.NumeroBloques = 2
						//Como se creará un usuario y un grupo en users.txt
						//se utilizaran aproximadamente 53 caracteres
						//cada bloque puede almacenar hasta 25 characeres, por lo tanto
						//se necesitarán 2 de los 4 bloques del inodo
						InodoAux.ApuntadoresBloques[0] = SB1.InicioBloques                  // 1,G,root\n1,U,root,root,20 <- caracteres en primer bloque
						InodoAux.ApuntadoresBloques[1] = SB1.InicioBloques + SB1.SizeBloque // 1602625 <- caracteres en segundo bloque

						//Ahora toca escribir el struct Inodo en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioInodos+1), 0)
						inodop := &InodoAux
						var binario5 bytes.Buffer
						binary.Write(&binario5, binary.BigEndian, inodop)
						escribirBytes(fileMBR, binario5.Bytes())

						//Bitmap de inodos (la primeras dos posiciones son para el archivo users.txt)
						//Escribiendo un 1 en las primeras 2 posiciones del bitmap
						fileMBR.Seek(int64(SB1.InicioBitmapBloques+1), 0)
						data = []byte{0x01}
						fileMBR.Write(data)
						fileMBR.Seek(int64(SB1.InicioBitmapBloques+2), 0)
						fileMBR.Write(data)

						//A continuación creamos el primer BloqueDatos
						BloqueAux := estructuras.BloqueDatos{}
						contenido := "1,G,root\n1,U,root,root,20"
						copy(BloqueAux.Data[:], contenido)

						//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
						fileMBR.Seek(int64(SB1.InicioBloques+1), 0)
						bloquep := &BloqueAux
						var binario6 bytes.Buffer
						binary.Write(&binario6, binary.BigEndian, bloquep)
						escribirBytes(fileMBR, binario6.Bytes())

						//A continuación creamos el segundo BloqueDatos
						BloqueAux2 := estructuras.BloqueDatos{}
						contenido = "1602625"
						copy(BloqueAux2.Data[:], contenido)
						//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
						fileMBR.Seek(int64((SB1.InicioBloques+SB1.SizeBloque)+1), 0)
						bloque2p := &BloqueAux2
						var binario7 bytes.Buffer
						binary.Write(&binario7, binary.BigEndian, bloque2p)
						escribirBytes(fileMBR, binario7.Bytes())

						color.Println("@{w} COMANDO LOSS realizadO con exito.")

					} else {
						color.Println("@{r} La partición indicada no ha sido formateada.")
					}

					fileMBR.Close()

				}

			} else {
				color.Printf("@{r}No hay ninguna partición montada con el id: @{w}%v\n", id)
			}

		} else {
			color.Println("@{r}Faltan parámetros obligatorios para la funcion LOSS.")
		}

	} else {
		color.Println("@{r}Se necesita de una sesión root activa para ejecutar la función LOSS.")
	}

}
