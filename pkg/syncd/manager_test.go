package syncd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_AllocateMutexes(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		hasToken func(m Manager, t string) bool
	}{
		{"inmem", ModeInMem, func(m Manager, t string) bool {
			_, ok := m.(*InMemManager).acqu[t]
			return ok
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewManager(tt.mode)
			assert.Nil(t, err)

			m.AllocateMutexes("X1", "X2", "X3")

			assert.True(t, tt.hasToken(m, "X1"))
			assert.True(t, tt.hasToken(m, "X2"))
			assert.True(t, tt.hasToken(m, "X3"))
		})
	}
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		mode        Mode
		correctType func(m Manager) bool
		wantErr     bool
	}{
		{"inmem", ModeInMem, func(m Manager) bool {
			_, ok := m.(*InMemManager)
			return ok
		}, false},
		{"invalid mode", -1, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewManager(tt.mode)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.True(t, tt.correctType(m))

			testMutexdFullFromManager(t, m)
			testRequestFullFromManager(t, m)
		})
	}
}

type errTestManager struct {
}

func (m *errTestManager) Lock(ctx context.Context, token string) error {
	return errors.New("lock")
}

func (m *errTestManager) RLock(ctx context.Context, token string) error {
	return errors.New("rlock")
}

func (m *errTestManager) Unlock(ctx context.Context, token string) error {
	return errors.New("unlock")
}

func (m *errTestManager) RUnlock(ctx context.Context, token string) error {
	return errors.New("runlock")
}

func (m *errTestManager) AllocateMutexes(tokens ...string) {}

func (m *errTestManager) NewRequest() *Request {
	return NewRequest(m)
}

func TestErrorManager(t *testing.T) {
	var m Manager = &errTestManager{}

	// Manager

	err := m.Lock(context.Background(), "")
	assert.NotNil(t, err)
	assert.Equal(t, "lock", err.Error())

	err = m.RLock(context.Background(), "")
	assert.NotNil(t, err)
	assert.Equal(t, "rlock", err.Error())

	err = m.Unlock(context.Background(), "")
	assert.NotNil(t, err)
	assert.Equal(t, "unlock", err.Error())

	err = m.RUnlock(context.Background(), "")
	assert.NotNil(t, err)
	assert.Equal(t, "runlock", err.Error())

	// Request

	r := m.NewRequest()

	err = r.Lock(context.Background(), "MAP_X110_Y40")
	assert.NotNil(t, err)
	assert.Equal(t, "lock", err.Error())
	_, ok := r.wLocks["MAP_X110_Y40"]
	assert.False(t, ok)

	err = r.RLock(context.Background(), "MAP_X110_Y40")
	assert.NotNil(t, err)
	assert.Equal(t, "rlock", err.Error())
	_, ok = r.rLocks["MAP_X110_Y40"]
	assert.False(t, ok)

	r.wLocks["MAP_X110_Y40"] = struct{}{}
	err = r.Unlock(context.Background(), "MAP_X110_Y40")
	assert.NotNil(t, err)
	assert.Equal(t, "unlock", err.Error())

	r.rLocks["MAP_X110_Y40"] = struct{}{}
	err = r.RUnlock(context.Background(), "MAP_X110_Y40")
	assert.NotNil(t, err)
	assert.Equal(t, "runlock", err.Error())

	// UnlockAll

	allErr := []string{}

	for err := range r.UnlockAll(context.Background()) {
		assert.NotNil(t, err)
		allErr = append(allErr, err.Error())
	}

	assert.Equal(t, []string{"runlock", "unlock"}, allErr)
}
