package workerpool

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMovingLater(t *testing.T) {
	inputs := make([]int, 1_000_000)
	var counter atomic.Int64
	for i := range inputs {
		inputs[i] = i
	}
	assert.LessOrEqual(
		t,
		MovingLater(
			inputs,
			func(i int) float64 {
				counter.Add(1)
				return float64(i)
			},
		),
		float64(inputs[len(inputs)-1]),
	)
	assert.NotZero(t, counter.Load())
}
