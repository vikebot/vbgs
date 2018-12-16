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
	testRequestAquiredCache(t, m.NewRequest())
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

	r.Lock(context.Background(), "MAP_X110_Y40")
	r.Lock(context.Background(), "MAP_X110_Y40")

	r.Unlock(context.Background(), "MAP_X110_Y40")
	r.Unlock(context.Background(), "MAP_X110_Y40")

	r.RLock(context.Background(), "MAP_X110_Y40")
	r.RLock(context.Background(), "MAP_X110_Y40")

	r.RUnlock(context.Background(), "MAP_X110_Y40")
	r.RUnlock(context.Background(), "MAP_X110_Y40")
}

func TestRequest_AquiredCache(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)

	testRequestAquiredCache(t, m.NewRequest())
}

func testRequestAquiredCache(t *testing.T, r *Request) {
	assert.NotNil(t, r)

	r.Lock(context.Background(), "MAP_X110_Y40")
	assert.True(t, r.Aquired("MAP_X110_Y40"))
	assert.False(t, r.Aquired("MAP"))
	r.Unlock(context.Background(), "MAP_X110_Y40")

	r.RLock(context.Background(), "MAP_X110_Y40")
	assert.True(t, r.RAquired("MAP_X110_Y40"))
	assert.False(t, r.RAquired("MAP"))
	r.RUnlock(context.Background(), "MAP_X110_Y40")
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

	chronology := []int{}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		r1.Lock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 0)
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		r2.Lock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 1)
		wg.Done()
	}()
	go func() {
		time.Sleep(2 * time.Second)
		r1.Unlock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 2)
		wg.Done()
	}()
	go func() {
		time.Sleep(3 * time.Second)
		r2.Unlock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 3)
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, []int{0, 2, 1, 3}, chronology)
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

	chronology := []int{}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		r1.RLock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 0)
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		r2.RLock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 1)
		wg.Done()
	}()
	go func() {
		time.Sleep(2 * time.Second)
		r1.RUnlock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 2)
		wg.Done()
	}()
	go func() {
		time.Sleep(3 * time.Second)
		r2.RUnlock(context.Background(), "MAP_X110_Y40")
		chronology = append(chronology, 3)
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, []int{0, 1, 2, 3}, chronology)
}

func TestRequest_UnlockAll(t *testing.T) {
	m, err := NewManager(ModeInMem)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	testRequestUnlockAll(t, m.NewRequest())
}

func testRequestUnlockAll(t *testing.T, r *Request) {
	assert.NotNil(t, r)

	r.Lock(context.Background(), "MAP1")
	r.Lock(context.Background(), "MAP2")
	r.RLock(context.Background(), "MAP3")
	r.RLock(context.Background(), "MAP4")

	err := r.UnlockAll(context.Background())
	for e := range err {
		assert.Nil(t, e)
	}

	r.Lock(context.Background(), "MAP1")
	r.Lock(context.Background(), "MAP2")
	r.RLock(context.Background(), "MAP3")
	r.RLock(context.Background(), "MAP4")
}
