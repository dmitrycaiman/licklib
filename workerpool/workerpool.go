package workerpool

import (
	"context"
	"sync"
)

// Workerpool запускает определёное количество исполнителей, которые конкурентно принимают данные на входном канале.
// При закрытии входного канала или отмены контекста все исполнители завершают работу.
func Workerpool[T1, T2 any](ctx context.Context, count int, input <-chan T1, f func(T1) T2) <-chan T2 {
	result := make(chan T2)
	var wg sync.WaitGroup

	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-input:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
					case result <- f(v):
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}
