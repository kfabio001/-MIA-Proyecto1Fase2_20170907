package estructuras

//Bitacora struct
type Bitacora struct {
	Contenido [300]byte
	Path      [300]byte
	Operacion [16]byte
	Proper    [16]byte
	Grupo     [16]byte
	Fecha     [20]byte
	Tipo      byte
	Size      int32
}
