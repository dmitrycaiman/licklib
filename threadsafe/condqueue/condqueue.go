package condqueue

import (
	"sync"
)

// CondQueue есть реализация многопоточной FIFO-очереди на Condition Variable.
type CondQueue[T any] struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	queue    []T
	capacity int
	closed   bool
}

func New[T any](capacity int) *CondQueue[T] {
	if capacity <= 0 {
		capacity = 1
	}
	output := &CondQueue[T]{
		queue:    make([]T, 0),
		capacity: capacity,
	}
	output.notEmpty = sync.NewCond(&output.mu)
	output.notFull = sync.NewCond(&output.mu)
	return output
}

func (slf *CondQueue[T]) Enqueue(value T) bool {
	slf.mu.Lock()
	defer slf.mu.Unlock()

	if slf.closed {
		return false
	}
	for len(slf.queue) >= slf.capacity {
		slf.notFull.Wait()
	}

	slf.queue = append(slf.queue, value)
	slf.notEmpty.Signal()

	return true
}

func (slf *CondQueue[T]) Dequeue() (T, bool) {
	slf.mu.Lock()
	defer slf.mu.Unlock()

	if slf.closed {
		return *new(T), false
	}
	for len(slf.queue) == 0 {
		slf.notEmpty.Wait()
	}

	value := slf.queue[0]
	slf.queue = slf.queue[1:]
	slf.notFull.Signal()

	return value, true
}

func (slf *CondQueue[T]) Close() {
	slf.mu.Lock()
	defer slf.mu.Unlock()
	slf.closed = true
}
