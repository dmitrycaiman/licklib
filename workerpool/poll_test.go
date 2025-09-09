package workerpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoll(t *testing.T) {
	var counter atomic.Int64
	f := func() (int64, error) {
		time.Sleep(10 * time.Millisecond)
		return counter.Add(1), nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-time.NewTimer(250 * time.Millisecond).C
		cancel()
	}()

	result := []int64{}
	for v := range Poll(ctx, 50*time.Millisecond, f) {
		result = append(result, v)
	}
	assert.ElementsMatch(t, []int64{1, 2, 3, 4}, result)
}
