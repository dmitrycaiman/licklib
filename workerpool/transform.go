package workerpool

import (
	"context"
	"sync"
)

// Transform перенаправляет из входного канала в выходной все значения, подвергая их трансформации.
func Transform[T1, T2 any](ctx context.Context, input <-chan T1, transform func(T1) T2, n int) <-chan T2 {
	output := make(chan T2)

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

					select {
					case output <- transform(value):
					case <-ctx.Done():
						return
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
