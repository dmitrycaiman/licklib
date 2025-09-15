package workerpool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Run(
		"general flow",
		func(t *testing.T) {
			input, result := make(chan int), []int{}
			output := Filter(context.Background(), input, func(n int) bool { return n%2 == 0 }, 10)

			go func() {
				defer close(input)
				for i := range 10_000 {
					input <- i
				}
			}()

			for v := range output {
				result = append(result, v)
			}

			assert.Len(t, result, 5_000)
			for _, v := range result {
				assert.True(t, v%2 == 0)
			}
		},
	)
	t.Run(
		"context cancel",
		func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			input, result := make(chan int), []int{}
			defer close(input)
			output := Filter(ctx, input, func(n int) bool { return n%2 == 0 }, 10)

			go func() { cancel() }()

			for v := range output {
				result = append(result, v)
			}

			assert.Empty(t, result)
		},
	)
	t.Run(
		"closed input channel",
		func(t *testing.T) {
			input, result := make(chan int), []int{}
			output := Filter(context.Background(), input, func(n int) bool { return n%2 == 0 }, 10)

			go func() {
				defer close(input)
				for i := range 10_000 {
					if i == 5_000 {
						return
					}
					input <- i
				}
			}()

			for v := range output {
				result = append(result, v)
			}

			assert.Len(t, result, 2_500)
			for _, v := range result {
				assert.True(t, v%2 == 0)
			}
		},
	)
}
