package meanval

import (
	"math"
	"sync"
)

const NonExistentValue = math.MinInt

// Meanval есть модуль расчёта и хранения средних значений.
type Meanval struct {
	mu      sync.RWMutex
	points  map[string]Resolver
	storage map[string]int
}

func NewMeanval() *Meanval { return &Meanval{points: map[string]Resolver{}, storage: map[string]int{}} }

// Upsert создаёт новую или заменяет существующую расчётную единицу.
func (slf *Meanval) Upsert(name string, engine Resolver) {
	slf.mu.Lock()
	defer slf.mu.Unlock()
	slf.points[name] = engine
}

// Update вызывает обновление значения расчётной единицы согласно новым данным.
func (slf *Meanval) Update(name string, value int) int {
	if name == "" {
		return 0
	}
	slf.mu.Lock()
	defer slf.mu.Unlock()
	point, ok := slf.points[name]
	if !ok {
		return NonExistentValue
	}
	newValue := point.Resolve(value)
	slf.storage[name] = newValue
	return newValue
}

// Select производит выдачу значения расчётной единицы.
func (slf *Meanval) Select(name string) int {
	slf.mu.RLock()
	defer slf.mu.RUnlock()
	val, ok := slf.storage[name]
	if !ok {
		return NonExistentValue
	}
	return val
}

// Reset производит сброс значения расчётной единицы.
// Входной аргумент в виде пустой строки приводит к сбросу значений всех расчётных единиц.
func (slf *Meanval) Reset(name string) {
	slf.mu.Lock()
	defer slf.mu.Unlock()
	if name == "" {
		slf.storage = map[string]int{}
	} else {
		delete(slf.storage, name)
	}
}

// Delete производит удаление расчётной единицы.
// Входной аргумент в виде пустой строки приводит к удалению всех расчётных единиц.
func (slf *Meanval) Delete(name string) {
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
