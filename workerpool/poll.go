package workerpool

import (
	"context"
	"sync"
	"time"
)

// Poll позволяет производить защищённый периодический запуск функции-запроса.
// Результаты запроса будут отправлены в выходной канал. Новый запрос не будет сделан, пока не закончен предыдущий.
func Poll[T any](ctx context.Context, period time.Duration, f func() (T, error)) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)

		type result struct {
			value T
			err   error
		}

		// Внутренний канал имеет буфер, чтобы не блокировалась горутина запроса.
		resultCh := make(chan *result, 1)

		ticker, busy := time.NewTicker(period), false
		defer ticker.Stop()

		// Дожидаемся завершения горутины запроса перед выходом.
		var wg sync.WaitGroup
		defer wg.Wait()

		for {
			select {
			case <-ticker.C:
				// Если не заняты обработкой предыдущего значения, то запускаем горутину запроса.
				if !busy {
					busy = true

					// Горутина позволяет избежать зависания на исполнении запроса.
					wg.Add(1)
					go func() {
						value, err := f()
						resultCh <- &result{value, err}
						wg.Done()
					}()
				}
			case result := <-resultCh:
				// Если произошла ошибка, не отправляем значение получателю.
				if result.err != nil {
					continue
				}

				// Производим отправку значения получателю с подстраховкой через отмену контекста.
				select {
				case output <- result.value:
					busy = false
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
