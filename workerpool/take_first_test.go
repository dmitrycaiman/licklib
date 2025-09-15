package workerpool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTakeFirstToList(t *testing.T) {
	t.Run(
		"general flow",
		func(t *testing.T) {
			input := make(chan int)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for i := range 1_000 {
					input <- i
				}
				close(input)
				wg.Done()
			}()

			assert.ElementsMatch(t, []int{0, 1, 2, 3, 4}, TakeFirstToList(context.Background(), 5, input))

			wg.Add(1)
			go func() {
				for range input {
				}
				wg.Done()
			}()

			wg.Wait()
		},
	)
	t.Run(
		"context cancel",
		func(t *testing.T) {
			input := make(chan int)
			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for i := range 1_000 {
					if i == 3 {
						cancel()
						time.Sleep(50 * time.Millisecond)
					}
					input <- i
				}
				close(input)
				wg.Done()
			}()

			assert.ElementsMatch(t, []int{0, 1, 2, 0, 0}, TakeFirstToList(ctx, 5, input))

			wg.Add(1)
			go func() {
				for range input {
				}
				wg.Done()
			}()

			wg.Wait()
		},
	)
	t.Run(
		"closed input channel",
		func(t *testing.T) {
			input := make(chan int)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for i := range 1_000 {
					if i == 3 {
						close(input)
						break
					}
					input <- i
				}
				wg.Done()
			}()

			assert.ElementsMatch(t, []int{0, 1, 2, 0, 0}, TakeFirstToList(context.Background(), 5, input))
			wg.Wait()
		},
	)
}

func TestTakeFirstToChan(t *testing.T) {
	t.Run(
		"general flow",
		func(t *testing.T) {
			input := make(chan int)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for i := range 1_000 {
					input <- i
				}
				close(input)
				wg.Done()
			}()

			result := []int{}
			for v := range TakeFirstToChan(context.Background(), 5, input) {
				result = append(result, v)
			}
			assert.ElementsMatch(t, []int{0, 1, 2, 3, 4}, result)

			wg.Add(1)
			go func() {
				for range input {
				}
				wg.Done()
			}()

			wg.Wait()
		},
	)
	t.Run(
		"context cancel",
		func(t *testing.T) {
			input := make(chan int)
			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for i := range 1_000 {
					if i == 3 {
						cancel()
						time.Sleep(50 * time.Millisecond)
					}
					input <- i
				}
				close(input)
				wg.Done()
			}()

			result := []int{}
			for v := range TakeFirstToChan(ctx, 5, input) {
				result = append(result, v)
			}
			assert.ElementsMatch(t, []int{0, 1, 2}, result)

			wg.Add(1)
			go func() {
				for range input {
				}
				wg.Done()
			}()

			wg.Wait()
		},
	)
	t.Run(
		"closed input channel",
		func(t *testing.T) {
			input := make(chan int)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for i := range 1_000 {
					if i == 3 {
						close(input)
						break
					}
					input <- i
				}
				wg.Done()
			}()

			result := []int{}
			for v := range TakeFirstToChan(context.Background(), 5, input) {
				result = append(result, v)
			}
			assert.ElementsMatch(t, []int{0, 1, 2}, result)

			wg.Wait()
		},
	)
}
