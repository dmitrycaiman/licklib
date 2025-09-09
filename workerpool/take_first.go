package workerpool

import "context"

// TakeFirstToList читает первые N значений из входного канала и возвращает их в виде слайса.
// Функция блокируется либо до отмены контекста, либо до закрытия входного канала.
func TakeFirstToList[T any](ctx context.Context, count int, input <-chan T) []T {
	result := make([]T, count)

loop:
	for i := range count {
		select {
		case value, ok := <-input:
			if !ok {
				break loop
			}
			result[i] = value
		case <-ctx.Done():
			break loop
		}
	}
	return result
}

// TakeFirstToChan посылает первые N значений из входного канала на выходной канал.
// Выходной канал будет закрыт либо по окончанию приёма, либо по отмене контекста, либо по закрытию входного канала.
func TakeFirstToChan[T any](ctx context.Context, count int, input <-chan T) <-chan T {
	output := make(chan T, count)

	go func() {
		defer close(output)

		for range count {
			select {
			case value, ok := <-input:
				if !ok {
					return
				}
				output <- value
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
