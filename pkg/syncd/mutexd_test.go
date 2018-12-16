package syncd

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testMutexdFullFromManager(t *testing.T, m Manager) {
	testMutexdLock(t, m)
	testMutexdRLock(t, m)
}

func testMutexdLock(t *testing.T, m Mutexd) {
	assert.NotNil(t, m)

	err := m.Lock(context.Background(), "MAP_X110_Y40")
	assert.Nil(t, err)

	var lastFinished int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		time.Sleep(2 * time.Second)
		m.Unlock(context.Background(), "MAP_X110_Y40")
		lastFinished = 0
		wg.Done()
	}()
	go func() {
		m.Lock(context.Background(), "MAP_X110_Y40")
		lastFinished = 1
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, 1, lastFinished)
}

func testMutexdRLock(t *testing.T, m Mutexd) {
	assert.NotNil(t, m)

	err := m.RLock(context.Background(), "MAP_X110_Y40")
	assert.Nil(t, err)

	var lastFinished int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		time.Sleep(2 * time.Second)
		m.RUnlock(context.Background(), "MAP_X110_Y40")
		lastFinished = 0
		wg.Done()
	}()
	go func() {
		m.RLock(context.Background(), "MAP_X110_Y40")
		lastFinished = 1
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, 0, lastFinished)
}
