package meanval

import "slices"

type median struct{ list []int }

func (slf *median) Resolve(value int) int {
	slf.list = append(slf.list, value)
	slices.Sort(slf.list)
	return slf.list[len(slf.list)/2]
}

// NewMedian создаёт расчётчик медианы.
func NewMedian() Resolver { return &median{[]int{}} }
