package estructuras

//MBR struct
type MBR struct {
	Msize       uint32
	Mdate       [20]byte
	Msignature  uint32
	Mpartitions [4]Partition
}
