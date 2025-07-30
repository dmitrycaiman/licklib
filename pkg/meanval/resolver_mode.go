package meanval

type mode struct{ counter map[int]int }

func (slf *mode) Resolve(value int) int {
	slf.counter[value]++
	max, output := -1, 0
	for k, v := range slf.counter {
		if v > max {
			output = k
		}
	}
	return output
}

// NewMode создаёт расчётчик моды.
func NewMode() Resolver { return &mode{map[int]int{}} }
