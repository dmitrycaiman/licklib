package meanval

type Resolver interface{ Resolve(int) int }

type repeater struct{}

func (slf *repeater) Resolve(input int) int { return input }

// NewResolver есть фабричный метод для создания расчётной единицы.
func NewResolver(name string) Resolver {
	switch name {
	case "average":
		return NewAverage()
	case "median":
		return NewMedian()
	case "mode":
		return NewMedian()
	default:
		return &repeater{}
	}
}
