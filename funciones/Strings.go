package funciones

import (
	"Proyecto1/estructuras"
	"bytes"
	"fmt"
)

//GenerarAVD devuelve un AVD seteado en formato string
func GenerarAVD(NoAVD int, AVDaux *estructuras.AVD) string {

	n := bytes.Index(AVDaux.NombreDir[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.NombreDir)
	}
	Directorio := string(AVDaux.NombreDir[:n])

	n = bytes.Index(AVDaux.FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.FechaCreacion)
	}
	Fecha := string(AVDaux.FechaCreacion[:n])

	n = bytes.Index(AVDaux.Proper[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.Proper)
	}
	Propietario := string(AVDaux.Proper[:n])

	n = bytes.Index(AVDaux.Grupo[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.Grupo)
	}
	Grupo := string(AVDaux.Grupo[:n])

	cadena := fmt.Sprintf(`AVD%v [label=<
	<TABLE BORDER="1"  cellpadding="2"   CELLBORDER="1" CELLSPACING="4" BGCOLOR="blue4" color = 'black'>            
	   <TR> 
		   <TD bgcolor='white' colspan="2"><font color='black' point-size='13'>Directorio: %s</font></TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='white' >Fecha creación</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >Propietario</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >Grupo</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >Permisos</TD>
		   <TD bgcolor='white' > %v%v%v </TD>
	   </TR>
	   
	   <TR>
		   <TD  bgcolor='white' >Apuntador Directorio</TD>
		   <TD  bgcolor='white' PORT="6"> %v </TD>
	   </TR>
	   <TR>
		   <TD  bgcolor='white' >Dir--</TD>
		   <TD  bgcolor='white' PORT="7"> %v</TD>
	   </TR>
   </TABLE>
	>];

	`, NoAVD, Directorio, Fecha, Propietario, Grupo, AVDaux.PermisoU, AVDaux.PermisoG, AVDaux.PermisoO, AVDaux.ApuntadorDD, AVDaux.ApuntadorAVD)

	return cadena
}

//GenerarDD devuelve un DD seteado en formato string
func GenerarDD(NoDD int, DDaux *estructuras.DD, carpeta string) string {

	// i = 0

	n := bytes.Index(DDaux.DDFiles[0].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[0].Name)
	}
	Nombre0 := string(DDaux.DDFiles[0].Name[:n])

	n = bytes.Index(DDaux.DDFiles[0].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[0].FechaCreacion)
	}
	Fc0 := string(DDaux.DDFiles[0].FechaCreacion[:n])

	n = bytes.Index(DDaux.DDFiles[0].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[0].FechaModificacion)
	}
	Fm0 := string(DDaux.DDFiles[0].FechaModificacion[:n])

	// i = 1

	n = bytes.Index(DDaux.DDFiles[1].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[1].Name)
	}

	n = bytes.Index(DDaux.DDFiles[1].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[1].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[1].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[1].FechaModificacion)
	}

	// i = 2

	n = bytes.Index(DDaux.DDFiles[2].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[2].Name)
	}

	n = bytes.Index(DDaux.DDFiles[2].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[2].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[2].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[2].FechaModificacion)
	}

	// i = 3

	n = bytes.Index(DDaux.DDFiles[3].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[3].Name)
	}

	n = bytes.Index(DDaux.DDFiles[3].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[3].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[3].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[3].FechaModificacion)
	}

	// i = 4

	n = bytes.Index(DDaux.DDFiles[4].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[4].Name)
	}

	n = bytes.Index(DDaux.DDFiles[4].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[4].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[4].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[4].FechaModificacion)
	}

	cadena := fmt.Sprintf(`DD%v [label=<
	<TABLE BORDER="1"  cellpadding="2"   CELLBORDER="1" CELLSPACING="4" BGCOLOR="blue4" color = 'black'>            
	   <TR> 
		   <TD bgcolor='white' colspan="2"><font color='black' point-size='13'>Detalles: %s</font></TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >[0].Nombre</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >[0].FechaCreacion</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
		<TR>
		   <TD bgcolor='white' >[0].FechaModificación</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
		<TR>
		   <TD bgcolor='white' >[0].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="0" > %v </TD>
	   </TR>
		
		<TR>
		   <TD bgcolor='white' >[1].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="1" > %v </TD>
	   </TR>
				<TR>
		   <TD bgcolor='white' >[2].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="2" > %v </TD>
	   </TR>
		<TR>
		   <TD bgcolor='white' >[3].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="3" > %v </TD>
	   </TR>
		
		<TR>
		   <TD bgcolor='white' >[4].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="4" > %v </TD>
	   </TR>
	   <TR>
		   <TD  bgcolor='white' >ApuntadorDD</TD>
		   <TD  bgcolor='white' PORT="5"> %v </TD>
	   </TR>

   </TABLE>
	>];

	`, NoDD, carpeta, Nombre0, Fc0, Fm0, DDaux.DDFiles[0].ApuntadorInodo, DDaux.DDFiles[1].ApuntadorInodo, DDaux.DDFiles[2].ApuntadorInodo, DDaux.DDFiles[3].ApuntadorInodo, DDaux.DDFiles[4].ApuntadorInodo, DDaux.ApuntadorDD)

	return cadena

}

//GenerarInodo devuelve un Inodo seteado en formato string
func GenerarInodo(NoInodo int, InodoAux *estructuras.Inodo) string {

	n := bytes.Index(InodoAux.Proper[:], []byte{0})
	if n == -1 {
		n = len(InodoAux.Proper)
	}
	Propietario := string(InodoAux.Proper[:n])

	n = bytes.Index(InodoAux.Grupo[:], []byte{0})
	if n == -1 {
		n = len(InodoAux.Grupo)
	}

	cadena := fmt.Sprintf(`Inodo%v [label=<
	<TABLE   cellpadding="2"   CELLBORDER="1" CELLSPACING="4" BGCOLOR="blue4" color = 'black'>            
	   <TR>
	   <TD bgcolor='dodgerblue' colspan="2"><font color='black' point-size='13'>Inodo %v</font></TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='deepskyblue' >Propietario</TD>
		   <TD bgcolor='deepskyblue' > %s </TD>
	   </TR>
	   
	   <TR> 
		   <TD bgcolor='deepskyblue' >File Size</TD>
		   <TD bgcolor='deepskyblue' > %v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='deepskyblue' >Numero de bloques</TD>
		   <TD bgcolor='deepskyblue' > %v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='deepskyblue' >Permisos</TD>
		   <TD bgcolor='deepskyblue' > %v%v%v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='mistyrose' >Ap[0]</TD>
		   <TD bgcolor='mistyrose' PORT="0" > %v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='mistyrose' >Ap[1]</TD>
		   <TD bgcolor='mistyrose' PORT="1" > %v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='mistyrose' >Ap[2]</TD>
		   <TD bgcolor='mistyrose' PORT="2" > %v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='mistyrose' >Ap[3]</TD>
		   <TD bgcolor='mistyrose' PORT="3" > %v </TD>
	   </TR>
	   <TR> 
		   <TD bgcolor='white' >ApuntadorIn</TD>
		   <TD bgcolor='white' PORT="4" > %v </TD>
	   </TR>

   	</TABLE>
   >];
   
	`, NoInodo, InodoAux.NumeroInodo, Propietario, InodoAux.FileSize, InodoAux.NumeroBloques, InodoAux.PermisoU, InodoAux.PermisoG, InodoAux.PermisoO, InodoAux.ApuntadoresBloques[0], InodoAux.ApuntadoresBloques[1], InodoAux.ApuntadoresBloques[2], InodoAux.ApuntadoresBloques[3], InodoAux.ApuntadorIndirecto)

	return cadena
}

//GenerarBloque devuelve un Bloque de datos seteado en formato string
func GenerarBloque(NoBloque int, Bloqueaux *estructuras.BloqueDatos) string {

	n := bytes.Index(Bloqueaux.Data[:], []byte{0})
	if n == -1 {
		n = len(Bloqueaux.Data)
	}
	contenido := string(Bloqueaux.Data[:n])

	cadena := fmt.Sprintf(`Bloque%v [label=<
	<table border="2" cellborder="0" cellspacing="1" bgcolor="lightsalmon" color="black">
		<tr> 
			<TD align ="center"><font color="white" >Bloque de Datos</font></TD> 
		</tr>
		<tr>
			<TD align="left"> %s </TD>
		</tr>
	</table>
	>];
	
	`, NoBloque, contenido)

	return cadena

}

//GenerarBitacora devuelve una bitacora seteada en formato string
func GenerarBitacora(NoBitacora int, BitacoraAux *estructuras.Bitacora) string {

	NumeroBitacora := NoBitacora + 1

	n := bytes.Index(BitacoraAux.Operacion[:], []byte{0})
	if n == -1 {
		n = len(BitacoraAux.Operacion)
	}
	Operacion := string(BitacoraAux.Operacion[:n])

	n = bytes.Index(BitacoraAux.Path[:], []byte{0})
	if n == -1 {
		n = len(BitacoraAux.Path)
	}
	Path := string(BitacoraAux.Path[:n])

	n = bytes.Index(BitacoraAux.Contenido[:], []byte{0})
	if n == -1 {
		n = len(BitacoraAux.Contenido)
	}
	Contenido := string(BitacoraAux.Contenido[:n])

	if n == 0 {
		Contenido = "-"
	}

	n = bytes.Index(BitacoraAux.Fecha[:], []byte{0})
	if n == -1 {
		n = len(BitacoraAux.Fecha)
	}
	Fecha := string(BitacoraAux.Fecha[:n])

	ValorTipo := "1"
	if BitacoraAux.Tipo != 1 {
		ValorTipo = "0"
	}

	//ValorSize := "-"
	//if BitacoraAux.Size != -1 {
	//ValorSize = string(BitacoraAux.Proper)
	//}
	s := bytes.Index(BitacoraAux.Proper[:], []byte{0})
	if s == -1 {
		s = len(BitacoraAux.Proper)
	}
	ValorSize := string(BitacoraAux.Proper[:s])

	cadena := fmt.Sprintf(`B%v [label=<
	<TABLE BORDER="1"  cellpadding="2"   CELLBORDER="1" CELLSPACING="4" BGCOLOR="blue4" color = 'black'>            
   	<TR> 
	   <TD bgcolor='white' colspan="2"><font color='black' point-size='13'>Journaling %v </font></TD>
   	</TR>
   	<TR>
	   <TD  bgcolor='white' > Operación </TD>
	   <TD  bgcolor='white'> %v </TD>
   	</TR>
   	<TR>
	   <TD  bgcolor='white' > Tipo </TD>
	   <TD  bgcolor='white' > %v </TD>
   	</TR>
   	<TR>
	   <TD  bgcolor='white' > Path </TD>
	   <TD  bgcolor='white' > %v </TD>
   	</TR>
   	<TR>
	   <TD  bgcolor='white' > Contenido </TD>
	   <TD  bgcolor='white' > %v </TD>
   	</TR>
   	<TR>
	   <TD  bgcolor='white' >Fecha Transa</TD>
	   <TD  bgcolor='white' > %v </TD>
   	</TR>
   	<TR>
	   <TD  bgcolor='white' >Propietario</TD>
	   <TD  bgcolor='white' > %v </TD>
   	</TR>

	</TABLE>
	>];
	
	`, NoBitacora, NumeroBitacora, Operacion, ValorTipo, Path, Contenido, Fecha, ValorSize)

	return cadena

}
func GenerarAVD3(NoAVD int, AVDaux *estructuras.AVD) string {

	n := bytes.Index(AVDaux.NombreDir[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.NombreDir)
	}

	n = bytes.Index(AVDaux.FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.FechaCreacion)
	}

	n = bytes.Index(AVDaux.Proper[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.Proper)
	}

	n = bytes.Index(AVDaux.Grupo[:], []byte{0})
	if n == -1 {
		n = len(AVDaux.Grupo)
	}

	cadena := fmt.Sprintf(`AVD%v [label=<
	<TABLE BORDER="1"  cellpadding="2"   CELLBORDER="1" CELLSPACING="4" BGCOLOR="blue4" color = 'black'>            
	   
	   
	   <TR>
		   <TD  bgcolor='white' >Apuntador Directorio</TD>
		   <TD  bgcolor='white' PORT="6"> %v </TD>
	   </TR>
	  
   </TABLE>
	>];

	`, NoAVD, AVDaux.ApuntadorDD)

	return cadena
}

//GenerarDD devuelve un DD seteado en formato string
func GenerarDD3(NoDD int, DDaux *estructuras.DD, carpeta string) string {

	// i = 0

	n := bytes.Index(DDaux.DDFiles[0].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[0].Name)
	}
	Nombre0 := string(DDaux.DDFiles[0].Name[:n])

	n = bytes.Index(DDaux.DDFiles[0].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[0].FechaCreacion)
	}
	Fc0 := string(DDaux.DDFiles[0].FechaCreacion[:n])

	n = bytes.Index(DDaux.DDFiles[0].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[0].FechaModificacion)
	}
	Fm0 := string(DDaux.DDFiles[0].FechaModificacion[:n])

	// i = 1

	n = bytes.Index(DDaux.DDFiles[1].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[1].Name)
	}

	n = bytes.Index(DDaux.DDFiles[1].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[1].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[1].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[1].FechaModificacion)
	}

	// i = 2

	n = bytes.Index(DDaux.DDFiles[2].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[2].Name)
	}

	n = bytes.Index(DDaux.DDFiles[2].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[2].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[2].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[2].FechaModificacion)
	}

	// i = 3

	n = bytes.Index(DDaux.DDFiles[3].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[3].Name)
	}

	n = bytes.Index(DDaux.DDFiles[3].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[3].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[3].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[3].FechaModificacion)
	}

	// i = 4

	n = bytes.Index(DDaux.DDFiles[4].Name[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[4].Name)
	}

	n = bytes.Index(DDaux.DDFiles[4].FechaCreacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[4].FechaCreacion)
	}

	n = bytes.Index(DDaux.DDFiles[4].FechaModificacion[:], []byte{0})
	if n == -1 {
		n = len(DDaux.DDFiles[4].FechaModificacion)
	}

	cadena := fmt.Sprintf(`DD%v [label=<
	<TABLE BORDER="1"  cellpadding="2"   CELLBORDER="1" CELLSPACING="4" BGCOLOR="blue4" color = 'black'>            
	   <TR> 
		   <TD bgcolor='white' colspan="2"><font color='black' point-size='13'>Detalles: %s</font></TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >[0].Nombre</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
	   <TR>
		   <TD bgcolor='white' >[0].FechaCreacion</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
		<TR>
		   <TD bgcolor='white' >[0].FechaModificación</TD>
		   <TD bgcolor='white' > %s </TD>
	   </TR>
		<TR>
		   <TD bgcolor='white' >[0].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="0" > %v </TD>
	   </TR>
		
		<TR>
		   <TD bgcolor='white' >[1].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="1" > %v </TD>
	   </TR>
				<TR>
		   <TD bgcolor='white' >[2].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="2" > %v </TD>
	   </TR>
		<TR>
		   <TD bgcolor='white' >[3].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="3" > %v </TD>
	   </TR>
		
		<TR>
		   <TD bgcolor='white' >[4].ApuntadorInodo</TD>
		   <TD bgcolor='white' PORT="4" > %v </TD>
	   </TR>
	   <TR>
		   <TD  bgcolor='white' >ApuntadorDD</TD>
		   <TD  bgcolor='white' PORT="5"> %v </TD>
	   </TR>

   </TABLE>
	>];

	`, NoDD, carpeta, Nombre0, Fc0, Fm0, DDaux.DDFiles[0].ApuntadorInodo, DDaux.DDFiles[1].ApuntadorInodo, DDaux.DDFiles[2].ApuntadorInodo, DDaux.DDFiles[3].ApuntadorInodo, DDaux.DDFiles[4].ApuntadorInodo, DDaux.ApuntadorDD)

	return cadena

}
