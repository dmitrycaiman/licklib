package meanval

import "sync"

// Meanval есть модуль расчёта и хранения средних значений.
type Meanval struct {
	mu      sync.RWMutex
	points  map[string]Resolver
	storage map[string]int
}

func NewMeanval() *Meanval { return &Meanval{points: map[string]Resolver{}, storage: map[string]int{}} }

func (slf *Meanval) Upsert(name string, engine Resolver) {
	slf.mu.Lock()
	defer slf.mu.Unlock()
	slf.points[name] = engine
	slf.storage[name] = 0
}

func (slf *Meanval) Update(name string, value int) int {
	if name == "" {
		return 0
	}
	slf.mu.Lock()
	defer slf.mu.Unlock()
	point, ok := slf.points[name]
	if !ok {
		return 0
	}
	newValue := point.Resolve(value)
	slf.storage[name] = newValue
	return newValue
}
func (slf *Meanval) Select(name string) int {
	slf.mu.RLock()
	defer slf.mu.RUnlock()
	return slf.storage[name]
}
func (slf *Meanval) Reset(name string) {
	slf.mu.Lock()
	defer slf.mu.Unlock()
	if name == "" {
		slf.points = map[string]Resolver{}
		slf.storage = map[string]int{}
	} else {
		delete(slf.points, name)
		delete(slf.storage, name)
	}
}
