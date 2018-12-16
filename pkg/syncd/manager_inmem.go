package syncd

import (
	"context"
	"sync"
)

// InMemManager implements Manager with in-memory sync.RWMutexes. Each token
// has it's own RWMutex
type InMemManager struct {
	aqu map[string]*sync.RWMutex
}

func newInMemManager() *InMemManager {
	return &InMemManager{
		aqu: make(map[string]*sync.RWMutex),
	}
}

// NewRequest returns a new Request based on the InMemManager
func (m *InMemManager) NewRequest() *Request {
	return newRequest(m)
}

// Lock locks the write-lock for the token if not already locked.
//
// Because InMemManager is a simple implementation of Manager, that doesn't
// rely on any dependencies or network calls, the passed context is never
// used. You are safe to use `context.Background()`. Furthermore there is
// no chance that an error is returned. The return value will always be nil.
func (m *InMemManager) Lock(_ context.Context, token string) error {
	if rw, ok := m.aqu[token]; ok {
		rw.Lock()
		return nil
	}

	var rw sync.RWMutex
	rw.Lock()

	m.aqu[token] = &rw

	return nil
}

// RLock locks the read-lock for the token if not already locked.
//
// Because InMemManager is a simple implementation of Manager, that doesn't
// rely on any dependencies or network calls, the passed context is never
// used. You are safe to use `context.Background()`. Furthermore there is
// no chance that an error is returned. The return value will always be nil.
func (m *InMemManager) RLock(_ context.Context, token string) error {
	if rw, ok := m.aqu[token]; ok {
		rw.RLock()
		return nil
	}

	var rw sync.RWMutex
	rw.RLock()

	m.aqu[token] = &rw

	return nil
}

// Unlock unlocks the write-lock for the token if not already unlocked.
//
// Because InMemManager is a simple implementation of Manager, that doesn't
// rely on any dependencies or network calls, the passed context is never
// used. You are safe to use `context.Background()`. Furthermore there is
// no chance that an error is returned. The return value will always be nil.
func (m *InMemManager) Unlock(_ context.Context, token string) error {
	if rw, ok := m.aqu[token]; ok {
		rw.Unlock()
		delete(m.aqu, token)
	}

	return nil
}

// RUnlock unlocks the read-lock for the token if not already unlocked.
//
// Because InMemManager is a simple implementation of Manager, that doesn't
// rely on any dependencies or network calls, the passed context is never
// used. You are safe to use `context.Background()`. Furthermore there is
// no chance that an error is returned. The return value will always be nil.
func (m *InMemManager) RUnlock(_ context.Context, token string) error {
	if rw, ok := m.aqu[token]; ok {
		rw.RUnlock()
		delete(m.aqu, token)
	}

	return nil
}
