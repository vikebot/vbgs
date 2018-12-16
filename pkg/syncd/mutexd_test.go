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

	assert.Nil(t, m.Lock(context.Background(), "MAP_X110_Y40"))

	chronology := make(chan int, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		assert.Nil(t, m.Lock(context.Background(), "MAP_X110_Y40"))
		chronology <- 0
		assert.Nil(t, m.Unlock(context.Background(), "MAP_X110_Y40"))
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		assert.Nil(t, m.Unlock(context.Background(), "MAP_X110_Y40"))
		chronology <- 1
		wg.Done()
	}()

	wg.Wait()
	close(chronology)

	chronologyArr := make([]int, 2)
	idx := 0
	for i := range chronology {
		chronologyArr[idx] = i
		idx++
	}

	assert.Equal(t, []int{1, 0}, chronologyArr)
}

func testMutexdRLock(t *testing.T, m Mutexd) {
	assert.NotNil(t, m)

	assert.Nil(t, m.RLock(context.Background(), "MAP_X110_Y40"))

	chronology := make(chan int, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		assert.Nil(t, m.RLock(context.Background(), "MAP_X110_Y40"))
		chronology <- 0
		assert.Nil(t, m.RUnlock(context.Background(), "MAP_X110_Y40"))
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		assert.Nil(t, m.RUnlock(context.Background(), "MAP_X110_Y40"))
		chronology <- 1
		wg.Done()
	}()

	wg.Wait()
	close(chronology)

	chronologyArr := make([]int, 2)
	idx := 0
	for i := range chronology {
		chronologyArr[idx] = i
		idx++
	}

	assert.Equal(t, []int{0, 1}, chronologyArr)
}
