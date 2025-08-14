package workerpool

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerpool(t *testing.T) {
	// Задаём функцию трансформации.
	f := func(i int) float64 { return float64(i * 2) }
	// Инициализируем входной канал пула.
	in := make(chan int)

	// Инициализируем пул исполнителей.
	out := Workerpool(context.Background(), 5, in, f)
	// Отдаём на исполнение входные данные. Здесь важно не превысить величину пула,
	// так как данные с выходного канала мы ещё не получаем — будет дедлок (никакая горутина не сможет принять входные данные).
	for i := range 5 {
		in <- i
	}

	// Обязательно завершаем работу с пулом через закрытие входного канала.
	// Данное действие приводит к закрытию выходного канала после получения всех результатов.
	// Если этого не сделать, то при сборе выходных данных будет дедлок (выходной канал останется открытым).
	assert.Equal(t, 6+2, runtime.NumGoroutine())
	close(in)

	// Собираем результаты работы функции трансформации из выходного канала. Порядок может отличаться от входного.
	result := []float64{}
	for v := range out {
		result = append(result, v)
		// Каждое получение результата приводит к завершению одной горутины, так как контекст отменён и входной канал закрыт.
		// После получения всех результатов спокойно выходим, так как выходной канал также в этот момент закроется.
	}

	// Проверяем результаты.
	assert.ElementsMatch(t, []float64{0, 2, 4, 6, 8}, result)
	assert.Equal(t, 0+2, runtime.NumGoroutine())

}

// BenchmarkWorkerpool показывает выигрыш в распараллеливании простых операций.
func BenchmarkWorkerpool(b *testing.B) {
	f := func(i int) float64 {
		time.Sleep(1 * time.Millisecond)
		return float64(i * 2)
	}
	n := 1_000
	b.Run(
		"with pool",
		func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				in := make(chan int, n)
				out := Workerpool(context.Background(), runtime.NumCPU(), in, f)
				for i := range n {
					in <- i
				}
				result := []float64{}
				close(in)
				for v := range out {
					result = append(result, v)
				}
				_ = result
			}
		},
	)
	b.Run(
		"no pool",
		func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				result := []float64{}
				for i := range n {
					result = append(result, f(i))
				}
				_ = result
			}
		},
	)
}
