package workerpool

import (
	"context"
	"sync"
)

// FanIn перенаправляет данные с нескольких входных каналов в один выходной канал.
// При закрытии всех входных каналов или отмены контекста будет закрыт выходной канал.
func FanIn[T any](ctx context.Context, input []<-chan T) <-chan T {
	result := make(chan T)
	var wg sync.WaitGroup

	for _, ch := range input {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case result <- v:
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

// FanOut перенаправляет данные из одного входного канала поочерёдно в несколько выходных каналов.
// При закрытии входного канала или отмены контекста будут закрыты выходные каналы.
func FanOut(ctx context.Context, input <-chan int, n int) []chan int {
	if n <= 0 {
		n = 1
	}

	output := make([]chan int, n)
	for i := range n {
		output[i] = make(chan int)
	}
	go func() {
		defer func() {
			for _, ch := range output {
				close(ch)
			}
		}()

		i := 0
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
					return
				case output[i] <- v:
					i = (i + 1) % n
				}
			}
		}
	}()

	return output
}
