package ge

type Location struct {
	X int64
	Y int64
}
type ARLocation struct {
	marshalToAbs bool
	Abs          Location
	Rel          Location
}
