package threadsafe

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemaphore(t *testing.T) {
	size, factor := 100, 10
	sem := NewSemaphore(size)
	var counter atomic.Int64

	var wg sync.WaitGroup
	for range size * factor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Aquire()
			counter.Add(1)
		}()
	}

	assert.Equal(t, int64(size), counter.Load())
	for counter.Load() != int64(size*factor) {
		for range rand.Intn(factor) {
			sem.Release()
		}
	}
	wg.Wait()
}
