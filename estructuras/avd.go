package estructuras

//AVD struct
type AVD struct {
	FechaCreacion [20]byte
	NombreDir     [20]byte
	Proper        [20]byte
	Grupo         [20]byte
	ApuntadorSubs [6]int32
	ApuntadorAVD  int32
	ApuntadorDD   int32
	PermisoU      int32
	PermisoG      int32
	PermisoO      int32
}
