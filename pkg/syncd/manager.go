package syncd

import (
	"errors"
)

// Manager defines the main interface a Distributed Lock Manager must provide
// in order to function correctly with syncd.
//
// It consists of the sub-interface Mutexd and a function for creating new
// child-managers called Requests. They can be used during a single request
// as they provide more failure-resistance than the plain manager itself.
type Manager interface {
	Mutexd
	AllocateMutexes(tokens ...string)
	NewRequest() *Request
}

// NewManager initializes a new Manager instance for the passed syncd.Mode
func NewManager(mode Mode) (c Manager, err error) {
	switch mode {
	case ModeInMem:
		return NewInMemManager(), nil
	}

	return nil, errors.New("syncd: unable to init manager without mode")
}
