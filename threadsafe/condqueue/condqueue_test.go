package condqueue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCondQueue(t *testing.T) {
	q, wg, signal := New[int](10), sync.WaitGroup{}, make(chan struct{})

	for i := range 100_000 {
		wg.Add(1)
		if i%2 == 0 {
			go func() {
				<-signal
				_, ok := q.Dequeue()
				assert.True(t, ok)
				wg.Done()
			}()
		} else {
			go func() {
				<-signal
				assert.True(t, q.Enqueue(i))
				wg.Done()
			}()
		}
	}
	close(signal)
	wg.Wait()
	assert.Empty(t, q.queue)

	q.Close()
	assert.False(t, q.Enqueue(0))
	_, ok := q.Dequeue()
	assert.False(t, ok)
}
