package workerpool

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSingleFlight(t *testing.T) {
	var counter int

	s := NewSingleFlight(
		func(key string) (int, error) {
			time.Sleep(100 * time.Millisecond)
			counter++
			return strconv.Atoi(key)
		},
	)

	var wg sync.WaitGroup
	for range 10_000 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result, err := s.Do(context.Background(), "12345")
			assert.Equal(t, result, 12345)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, 1, counter)
}
