package workerpool

import (
	"context"
	"sync"
)

// Filter перенаправляет из входного канала в выходной только те значения, которые соответствуют предикату.
func Filter[T any](ctx context.Context, input <-chan T, preicate func(T) bool, n int) <-chan T {
	output := make(chan T)

	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case value, ok := <-input:
					if !ok {
						return
					}

					if preicate(value) {
						select {
						case output <- value:
						case <-ctx.Done():
							return
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
