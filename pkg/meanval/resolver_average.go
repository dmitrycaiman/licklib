package meanval

type average struct{ sum, count int }

func (slf *average) Resolve(value int) int {
	slf.sum += value
	slf.count++
	return slf.sum / slf.count
}

// NewAverage создаёт расчётчик среднего арифметического.
func NewAverage() Resolver { return &average{} }
