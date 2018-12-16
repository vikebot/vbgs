package syncd

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testRequestFullFromManager(t *testing.T, m Manager) {
	assert.NotNil(t, m)

	testRequestMultipleCalls(t, m.NewRequest())
	testRequestAcquiredCache(t, m.NewRequest())
	testRequestLockChronology(t, m.NewRequest(), m.NewRequest())
	testRequestRLockChronology(t, m.NewRequest(), m.NewRequest())
	testRequestUnlockAll(t, m.NewRequest())
}

func TestRequest_MultipleCalls(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)

	testRequestMultipleCalls(t, m.NewRequest())
}

func testRequestMultipleCalls(t *testing.T, r *Request) {
	assert.NotNil(t, r)

	assert.Nil(t, r.Lock(context.Background(), "MAP_X110_Y40"))
	assert.Nil(t, r.Lock(context.Background(), "MAP_X110_Y40"))

	assert.Nil(t, r.Unlock(context.Background(), "MAP_X110_Y40"))
	assert.Nil(t, r.Unlock(context.Background(), "MAP_X110_Y40"))

	assert.Nil(t, r.RLock(context.Background(), "MAP_X110_Y40"))
	assert.Nil(t, r.RLock(context.Background(), "MAP_X110_Y40"))

	assert.Nil(t, r.RUnlock(context.Background(), "MAP_X110_Y40"))
	assert.Nil(t, r.RUnlock(context.Background(), "MAP_X110_Y40"))
}

func TestRequest_AcquiredCache(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)

	testRequestAcquiredCache(t, m.NewRequest())
}

func testRequestAcquiredCache(t *testing.T, r *Request) {
	assert.NotNil(t, r)

	assert.Nil(t, r.Lock(context.Background(), "MAP_X110_Y40"))
	assert.True(t, r.Acquired("MAP_X110_Y40"))
	assert.False(t, r.Acquired("MAP"))
	assert.Nil(t, r.Unlock(context.Background(), "MAP_X110_Y40"))

	assert.Nil(t, r.RLock(context.Background(), "MAP_X110_Y40"))
	assert.True(t, r.RAcquired("MAP_X110_Y40"))
	assert.False(t, r.RAcquired("MAP"))
	assert.Nil(t, r.RUnlock(context.Background(), "MAP_X110_Y40"))
}

func TestRequest_LockChronology(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	testRequestLockChronology(t, m.NewRequest(), m.NewRequest())
}

func testRequestLockChronology(t *testing.T, r1, r2 *Request) {
	assert.NotNil(t, r1)
	assert.NotNil(t, r2)

	chronology := make(chan int, 4)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		assert.Nil(t, r1.Lock(context.Background(), "MAP_X110_Y40"))
		chronology <- 0

		time.Sleep(2 * time.Second)

		assert.Nil(t, r1.Unlock(context.Background(), "MAP_X110_Y40"))
		chronology <- 2

		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)

		assert.Nil(t, r2.Lock(context.Background(), "MAP_X110_Y40"))
		chronology <- 1

		time.Sleep(2 * time.Second)

		assert.Nil(t, r2.Unlock(context.Background(), "MAP_X110_Y40"))
		chronology <- 3

		wg.Done()
	}()

	wg.Wait()
	close(chronology)

	chronologyArr := make([]int, 4)
	idx := 0
	for i := range chronology {
		chronologyArr[idx] = i
		idx++
	}

	assert.Equal(t, []int{0, 2, 1, 3}, chronologyArr)
}

func TestRequest_RLockChronology(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	testRequestRLockChronology(t, m.NewRequest(), m.NewRequest())
}

func testRequestRLockChronology(t *testing.T, r1, r2 *Request) {
	assert.NotNil(t, r1)
	assert.NotNil(t, r2)

	chronology := make(chan int, 4)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		assert.Nil(t, r1.RLock(context.Background(), "MAP_X110_Y40"))
		chronology <- 0

		time.Sleep(2 * time.Second)

		assert.Nil(t, r1.RUnlock(context.Background(), "MAP_X110_Y40"))
		chronology <- 2

		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)

		assert.Nil(t, r2.RLock(context.Background(), "MAP_X110_Y40"))
		chronology <- 1

		time.Sleep(2 * time.Second)

		assert.Nil(t, r2.RUnlock(context.Background(), "MAP_X110_Y40"))
		chronology <- 3

		wg.Done()
	}()

	wg.Wait()
	close(chronology)

	chronologyArr := make([]int, 4)
	idx := 0
	for i := range chronology {
		chronologyArr[idx] = i
		idx++
	}

	assert.Equal(t, []int{0, 1, 2, 3}, chronologyArr)
}

func TestRequest_UnlockAll(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	testRequestUnlockAll(t, m.NewRequest())
}

func testRequestUnlockAll(t *testing.T, r *Request) {
	assert.NotNil(t, r)

	assert.Nil(t, r.Lock(context.Background(), "MAP1"))
	assert.Nil(t, r.Lock(context.Background(), "MAP2"))
	assert.Nil(t, r.RLock(context.Background(), "MAP3"))
	assert.Nil(t, r.RLock(context.Background(), "MAP4"))

	err := r.UnlockAll(context.Background())
	for e := range err {
		assert.Nil(t, e)
	}

	assert.Nil(t, r.Lock(context.Background(), "MAP1"))
	assert.Nil(t, r.Lock(context.Background(), "MAP2"))
	assert.Nil(t, r.RLock(context.Background(), "MAP3"))
	assert.Nil(t, r.RLock(context.Background(), "MAP4"))
}
