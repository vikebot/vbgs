package syncd

// Mode is the mode of a Manager instance. For example 'InMem' or 'Etcd'.
type Mode int

const (
	// InMem is the in-memory mode of a Manager. It consists of a in-memory
	// hashmap of token:Mutex pairs.
	InMem Mode = iota
)
