package estructuras

//EBR struct
type EBR struct {
	Ename   [16]byte
	Estart  uint32
	Esize   uint32
	Enext   int32
	Eprev   int32
	Estatus byte
	Efit    byte
}
