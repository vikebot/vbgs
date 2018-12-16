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

	chronology := make(chan int, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		m.Lock(context.Background(), "MAP_X110_Y40")
		chronology <- 0
		m.Unlock(context.Background(), "MAP_X110_Y40")
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		m.Unlock(context.Background(), "MAP_X110_Y40")
		chronology <- 1
		wg.Done()
	}()

	wg.Wait()
	close(chronology)

	var chronologyArr []int
	for i := range chronology {
		chronologyArr = append(chronologyArr, i)
	}

	assert.Equal(t, []int{1, 0}, chronologyArr)
}

func testMutexdRLock(t *testing.T, m Mutexd) {
	assert.NotNil(t, m)

	err := m.RLock(context.Background(), "MAP_X110_Y40")
	assert.Nil(t, err)

	chronology := make(chan int, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		m.RLock(context.Background(), "MAP_X110_Y40")
		chronology <- 0
		m.RUnlock(context.Background(), "MAP_X110_Y40")
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		m.RUnlock(context.Background(), "MAP_X110_Y40")
		chronology <- 1
		wg.Done()
	}()

	wg.Wait()
	close(chronology)

	var chronologyArr []int
	for i := range chronology {
		chronologyArr = append(chronologyArr, i)
	}

	assert.Equal(t, []int{0, 1}, chronologyArr)
}
