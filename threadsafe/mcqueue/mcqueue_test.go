package mcqueue

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCondQueue(t *testing.T) {
	q, wg, signal := New[int64](), sync.WaitGroup{}, make(chan struct{})
	var result atomic.Int64

	for i := range 100_000 {
		wg.Add(1)
		if i%2 == 0 {
			go func() {
				<-signal
				popValue, ok := q.Dequeue()
				assert.True(t, ok)
				result.Add(-popValue)
				wg.Done()
			}()
		} else {
			go func() {
				pushValue := rand.Int63n(1_000)
				result.Add(pushValue)

				<-signal
				assert.True(t, q.Enqueue(pushValue))
				wg.Done()
			}()
		}
	}
	close(signal)
	wg.Wait()

	assert.Zero(t, result.Load())

	q.Close()
	assert.False(t, q.Enqueue(0))
	_, ok := q.Dequeue()
	assert.False(t, ok)
}
