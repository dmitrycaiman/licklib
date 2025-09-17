package workerpool

import "sync"

// MovingLater отправляет на запуск несколько экземпляров одной и той же функции с разными входными данными
// и возвращает первый полученный результат.
func MovingLater[T1, T2 any](inputs []T1, f func(T1) T2) T2 {
	output := make(chan T2, 1)

	var wg sync.WaitGroup
	for _, input := range inputs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case output <- f(input):
			default:
			}
		}()
	}
	go func() {
		wg.Wait()
		close(output)
	}()

	return <-output
}
