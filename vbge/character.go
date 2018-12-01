package vbge

type character interface {
	Rotate() NotifyGroup
	Move() (NotifyGroup, error)
	Radar() (int, NotifyGroup, error)
	Scout() (int, NotifyGroup, error)
	Environment() ([][]string, NotifyGroup, error)
	Watch() ([][]int, NotifyGroup, error)
	Attack() (int, NotifyGroup, error)
	Defend() (NotifyGroup, error)
	Undefend() (NotifyGroup, error)
}
