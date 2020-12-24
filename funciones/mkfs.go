package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/doun/terminal/color"
)

//EjecutarMkfs inicia el formateo de una particion
func EjecutarMkfs(id string, tipo string, add string, unit string) {

	if id != "" {

		if IDYaRegistrado(id) {

			color.Println("@{w}Listo para montar sistema de archivos")

			NameAux, PathAux := GetDatosPart(id)

			if Existe, Indice := ExisteParticion(PathAux, NameAux); Existe { //SIGNIFICA QUE ES PRIMARIA PORQUE EN LA FUNCION MOUNT SE VERIFICA QUE NO SEA EXTENDIDA ANTES DE MONTAR.

				PartStart, PartSize := GetStartAndSize(PathAux, Indice)

				color.Printf("@{w}La particion inicia en el: %d byte\n ", PartStart)
				color.Printf("@{w}La particion tiene un size de: %d\n", PartSize)

				Formatear(PartStart, PartSize, tipo, PathAux)

			} else if ExisteL, IndiceL := ExisteParticionLogica(PathAux, NameAux); ExisteL { //SIGNIFICA QUE ES LOGICA (NO ES REQUISITO)

				fileMBR, err := os.OpenFile(PathAux, os.O_RDWR, 0666)
				if err != nil {
					fmt.Println(err)
					fileMBR.Close()
				}

				//EBRaux sera el apuntador al struct EBR temporal
				EBRAux := estructuras.EBR{}
				EBRSize := int(unsafe.Sizeof(EBRAux))
				InicioParticion := IndiceL + EBRSize
				fileMBR.Seek(int64(InicioParticion+1), 0)
				EBRData := leerBytes(fileMBR, EBRSize)
				buffer := bytes.NewBuffer(EBRData)
				err = binary.Read(buffer, binary.BigEndian, &EBRAux)
				if err != nil {
					fileMBR.Close()
					panic(err)
				}

				fileMBR.Close()

				color.Printf("@{w}La particion inicia en el byte: %d\n", InicioParticion)
				color.Printf("@{w}La particion tiene un size de: %d\n", EBRAux.Esize)

				Formatear(InicioParticion, int(EBRAux.Esize), tipo, PathAux)

			}

		} else {
			color.Printf("@{r}No hay ninguna partición montada con el id: %v\n", id)
		}

	} else {
		color.Println("@{r}Falta parámetro -id, obligatorio para ejecutar la función Mkfs.")
	}

}

//Formatear ejecuta el formateo Full O Fast según se requiera
func Formatear(PartStart int, PartSize int, tipo string, path string) {

	sizeSuperbloque := int32(unsafe.Sizeof(estructuras.Superblock{}))
	startPart := int32(PartStart)
	sizePart := int32(PartSize)

	if (sizePart - (2 * sizeSuperbloque)) > 0 {

		//Obteniendo sizes de cada structs

		sizeAVD := int32(unsafe.Sizeof(estructuras.AVD{}))
		sizeDD := int32(unsafe.Sizeof(estructuras.DD{}))
		sizeInodo := int32(unsafe.Sizeof(estructuras.Inodo{}))
		sizeBloque := int32(unsafe.Sizeof(estructuras.BloqueDatos{}))
		sizeBitacora := int32(unsafe.Sizeof(estructuras.Bitacora{}))

		//Calculando número de estructuras

		NumEstructuras := (sizePart - (2 * sizeSuperbloque)) / (27 + sizeAVD + sizeDD + (5*sizeInodo + (20 * sizeBloque) + sizeBitacora))

		cantidadAVDS := NumEstructuras
		cantidadDDS := NumEstructuras
		cantidadInodos := int32(5 * NumEstructuras)
		cantidadBloques := int32(4 * cantidadInodos)
		cantidadBitacoras := NumEstructuras

		//Seteando Superbloque

		sb := estructuras.Superblock{}
		sb.PartStart = int32(PartStart)
		sb.PartSize = int32(PartSize)
		var chars [16]byte
		VirtualName := filepath.Base(path)
		copy(chars[:], VirtualName)
		copy(sb.Name[:], chars[:])
		sb.TotalAVDS = cantidadAVDS
		sb.TotalDDS = cantidadDDS
		sb.TotalInodos = cantidadInodos
		sb.TotalBloques = cantidadBloques
		sb.TotalBitacoras = cantidadBitacoras
		sb.FreeAVDS = cantidadAVDS - int32(1)
		sb.FreeDDS = cantidadDDS - int32(1)
		sb.FreeInodos = cantidadInodos - int32(1)       //Restamos 1 bloque porque al escribir users.txt ocupamos 1 bloque
		sb.FreeBloques = cantidadBloques - int32(2)     //Restamos 2 bloques porque al escribir users.txt ocupamos 2 bloques
		sb.FreeBitacoras = cantidadBitacoras - int32(2) //Restamos porque al crear la carpeta Root y el archivo Users.txt ocupamos 2 bitacoras
		t := time.Now()
		var charsDate [20]byte
		cadena := t.Format("2006-01-02 15:04:05")
		copy(charsDate[:], cadena)
		copy(sb.DateCreacion[:], charsDate[:])
		copy(sb.DateLastMount[:], charsDate[:])
		sb.MontajesCount = 1
		sb.InicioBitmapAVDS = startPart + sizeSuperbloque
		sb.InicioAVDS = sb.InicioBitmapAVDS + cantidadAVDS
		sb.InicioBitMapDDS = sb.InicioAVDS + (sizeAVD * cantidadAVDS)
		sb.InicioDDS = sb.InicioBitMapDDS + cantidadDDS
		sb.InicioBitmapInodos = sb.InicioDDS + (sizeDD * cantidadDDS)
		sb.InicioInodos = sb.InicioBitmapInodos + cantidadInodos
		sb.InicioBitmapBloques = sb.InicioInodos + (sizeInodo * cantidadInodos)
		sb.InicioBloques = sb.InicioBitmapBloques + cantidadBloques
		sb.InicioBitacora = sb.InicioBloques + (sizeBloque * cantidadBloques)
		sb.SizeAVD = sizeAVD
		sb.SizeDD = sizeDD
		sb.SizeInodo = sizeInodo
		sb.SizeBloque = sizeBloque
		sb.SizeBitacora = sizeBitacora
		sb.FirstFreeAVD = sb.InicioAVDS + sb.SizeAVD                //Le sumamos un sizeAVD porque vamos a crear la carpeta "/"
		sb.FirstFreeDD = sb.InicioDDS + sb.SizeDD                   //Le sumamos un sizeDD porque vamos a crear el DD de la carpeta "/"
		sb.FirstFreeInodo = sb.InicioInodos + sb.SizeInodo          //Le sumamos un sizeInodo porque vamos a crear el inodo del archivo users.txt
		sb.FirstFreeBloque = sb.InicioBloques + int32(2*sizeBloque) //Le sumamos dos sizeBloque porque, vamos a crear un usuario y un grupo default, lo cual abarca 32 caracteres.
		sb.MagicNum = 123

		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
			file.Close()
		}

		//////LEEMOS UN SB PARA OBTENER EL ATRIBUTO Montajes count
		file.Seek(int64(PartStart+1), 0)
		SBtemp := estructuras.Superblock{}
		SBtam := int(unsafe.Sizeof(SBtemp))
		SBD := leerBytes(file, SBtam)
		bufferT := bytes.NewBuffer(SBD)
		err = binary.Read(bufferT, binary.BigEndian, &SBtemp)
		if err != nil {
			file.Close()
			fmt.Println(err)
			return
		}

		if SBtemp.MontajesCount > 0 {
			copy(sb.DateCreacion[:], SBtemp.DateCreacion[:])
		}

		//LE SUMAMOS 1 A LA CANTIDAD ANTERIOR
		sb.MontajesCount += SBtemp.MontajesCount
		color.Printf("@{w}El conteo de formatos para la particion es @{w}%v @{w}formateo(s).\n", int(sb.MontajesCount))
		////////////////////////////////

		file.Seek(int64(PartStart+1), 0)
		if strings.ToLower(tipo) == "full" {
			data := make([]byte, sizePart)
			file.Write(data)
			file.Seek(int64(PartStart+1), 0)
		}
		//Escribiendo el Superbloque
		sb1 := &sb
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, sb1)
		escribirBytes(file, binario1.Bytes())

		//Escribiendo el Bitmap de AVDS
		file.Seek(int64(sb.InicioBitmapAVDS+1), 0)
		data := make([]byte, cantidadAVDS)
		file.Write(data)

		//Escribiendo el Bitmap de DDS
		file.Seek(int64(sb.InicioBitMapDDS+1), 0)
		data = make([]byte, cantidadDDS)
		file.Write(data)

		//Escribiendo de Bitmap de Inodos
		file.Seek(int64(sb.InicioBitmapInodos+1), 0)
		data = make([]byte, cantidadInodos)
		file.Write(data)

		//Escribiendo el Bitmap de bloques
		file.Seek(int64(sb.InicioBitmapBloques+1), 0)
		data = make([]byte, cantidadBloques)
		file.Write(data)

		//Escribir el Backup del Superbloque
		file.Seek(int64((sb.InicioBitacora+(sizeBitacora*cantidadBitacoras))+1), 0)
		sb2 := &sb
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, sb2)
		escribirBytes(file, binario2.Bytes())

		//Creando folder / y users.txt
		//Bitmap de AVD (la primera posición es la carpeta "/")
		//Escribiendo un 1 en la primera posicion del bitmap
		file.Seek(int64(sb.InicioBitmapAVDS+1), 0)
		data = []byte{0x01}
		file.Write(data)
		//Seteando valores de la carpeta root "/
		//Creamos nueva estructura AVD, la cual será escrita en su posición correspondiente
		AVDaux := estructuras.AVD{}
		t = time.Now()
		cadena = t.Format("2006-01-02 15:04:05")
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
		AVDaux.ApuntadorDD = sb.InicioDDS
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
		file.Seek(int64(sb.InicioAVDS+1), 0)
		avdp := &AVDaux
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, avdp)
		escribirBytes(file, binario3.Bytes())

		//Bitmap de DD (la primera posición es el DD de la carpeta "/")
		//Escribiendo un 1 en la primera posicion del bitmap
		file.Seek(int64(sb.InicioBitMapDDS+1), 0)
		data = []byte{0x01}
		file.Write(data)
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
		DDaux.DDFiles[0].ApuntadorInodo = sb.InicioInodos

		//Ahora toca escribir el struct DD en su posición correspondiente
		file.Seek(int64(sb.InicioDDS+1), 0)
		ddp := &DDaux
		var binario4 bytes.Buffer
		binary.Write(&binario4, binary.BigEndian, ddp)
		escribirBytes(file, binario4.Bytes())

		//Bitmap de inodos (la primera posición es el inodo para el archivo users.txt)
		//Escribiendo un 1 en la primera posicion del bitmap
		file.Seek(int64(sb.InicioBitmapInodos+1), 0)
		data = []byte{0x01}
		file.Write(data)

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
		InodoAux.ApuntadoresBloques[0] = sb.InicioBloques              // 1,G,root\n1,U,root,root,20 <- caracteres en primer bloque
		InodoAux.ApuntadoresBloques[1] = sb.InicioBloques + sizeBloque // 1602625 <- caracteres en segundo bloque

		//Ahora toca escribir el struct Inodo en su posición correspondiente
		file.Seek(int64(sb.InicioInodos+1), 0)
		inodop := &InodoAux
		var binario5 bytes.Buffer
		binary.Write(&binario5, binary.BigEndian, inodop)
		escribirBytes(file, binario5.Bytes())

		//Bitmap de inodos (la primeras dos posiciones son para el archivo users.txt)
		//Escribiendo un 1 en las primeras 2 posiciones del bitmap
		file.Seek(int64(sb.InicioBitmapBloques+1), 0)
		data = []byte{0x01}
		file.Write(data)
		file.Seek(int64(sb.InicioBitmapBloques+2), 0)
		file.Write(data)

		//A continuación creamos el primer BloqueDatos
		BloqueAux := estructuras.BloqueDatos{}
		contenido := "1,G,root\n1,U,root,root,12"
		copy(BloqueAux.Data[:], contenido)

		//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
		file.Seek(int64(sb.InicioBloques+1), 0)
		bloquep := &BloqueAux
		var binario6 bytes.Buffer
		binary.Write(&binario6, binary.BigEndian, bloquep)
		escribirBytes(file, binario6.Bytes())

		//A continuación creamos el segundo BloqueDatos
		BloqueAux2 := estructuras.BloqueDatos{}
		contenido = "3______"
		copy(BloqueAux2.Data[:], contenido)
		//Ahora toca escribir el struct BloqueDatos en su posición correspondiente
		file.Seek(int64((sb.InicioBloques+sizeBloque)+1), 0)
		bloque2p := &BloqueAux2
		var binario7 bytes.Buffer
		binary.Write(&binario7, binary.BigEndian, bloque2p)
		escribirBytes(file, binario7.Bytes())

		///////////// ESCRIBIENDO BITACORAS ////////////////////////
		//Creamos la bitacora para la creación de la carpeta raiz "/"
		BitacoraAux := estructuras.Bitacora{}
		//Seteamos el path, en este caso la primera carpeta tiene "/" como path
		var PathChars [300]byte
		PathAux := "/"
		copy(PathChars[:], PathAux)
		copy(BitacoraAux.Path[:], PathChars[:])
		//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
		var OperacionChars [16]byte
		OperacionAux := "Mkdir"
		copy(OperacionChars[:], OperacionAux)
		copy(BitacoraAux.Operacion[:], OperacionChars[:])
		//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
		BitacoraAux.Tipo = 1
		BitacoraAux.Size = -1
		//Seteamo la fecha de creación de la bitácora
		t = time.Now()
		var charsTime [20]byte
		cadena = t.Format("2006-01-02 15:04:05")
		copy(charsTime[:], cadena)
		copy(BitacoraAux.Fecha[:], charsTime[:])
		//Calculamos la posicion en la particion donde debemos escribir la bitacora
		//en este caso al ser la primera bitacora ira al inicio del bloque de bitacoras
		BitacoraPos := sb.InicioBitacora
		//Ahora toca escribir el struct Bitacora en su posición correspondiente
		file.Seek(int64(BitacoraPos+1), 0)
		bitacorap := &BitacoraAux
		var binario8 bytes.Buffer
		binary.Write(&binario8, binary.BigEndian, bitacorap)
		escribirBytes(file, binario8.Bytes())

		//Creamos la bitacora para la creación del archivo "/users.txt"
		BitacoraAux2 := estructuras.Bitacora{}
		//Seteamos el path, en este caso el archivo tiene "/users.txt"
		var PathChars2 [300]byte
		PathAux2 := "/users.txt"
		copy(PathChars2[:], PathAux2)
		copy(BitacoraAux2.Path[:], PathChars2[:])
		//Seteamos el nombre de la operacion encargada de crear carpetas "Mkdir"
		var OperacionChars2 [16]byte
		OperacionAux2 := "Mkfile"
		copy(OperacionChars2[:], OperacionAux2)
		copy(BitacoraAux2.Operacion[:], OperacionChars2[:])
		//Seteamos el contenido del archivo users.txt
		var ContenidoChars [300]byte
		contenidoAux := "1,G,root\n1,U,root,root,123"
		copy(ContenidoChars[:], contenidoAux)
		copy(BitacoraAux2.Contenido[:], ContenidoChars[:])
		//Seteamos el tipo con un 1 (1 significa carpeta, 2 significa archivo)
		BitacoraAux2.Tipo = 0
		BitacoraAux2.Size = 32
		//Seteamos la fecha de creación de la bitácora
		t = time.Now()
		var charsTime2 [20]byte
		cadena = t.Format("2006-01-02 15:04:05")
		copy(charsTime2[:], cadena)
		copy(BitacoraAux2.Fecha[:], charsTime2[:])
		//Calculamos la posicion en la particion donde debemos escribir la bitacora
		//en este caso será la segunda bitacora
		BitacoraPos = sb.InicioBitacora + sizeBitacora
		//Ahora toca escribir el struct Bitacora en su posición correspondiente
		file.Seek(int64(BitacoraPos+1), 0)
		bitacorap2 := &BitacoraAux2
		var binario9 bytes.Buffer
		binary.Write(&binario9, binary.BigEndian, bitacorap2)
		escribirBytes(file, binario9.Bytes())

		file.Close()

		sesionActiva = false
		sesionRoot = false
		idSesion = ""
		idGrupo = ""

		color.Println("@{w}Sistema de archivos fue instalado.")

	} else {
		color.Println("@{r}El tamaño de la partición es insuficiente para montar el sistema.")
	}
}
