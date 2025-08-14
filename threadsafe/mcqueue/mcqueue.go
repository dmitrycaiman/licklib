package mcqueue

import (
	"sync/atomic"
)

type item[T any] struct {
	value T
	next  atomic.Pointer[item[T]]
}

// CondQueue есть реализация многопоточной FIFO-очереди Майкла-Скотта (lockfree).
type MichaelScottQueue[T any] struct {
	head, tail atomic.Pointer[item[T]]
	closed     atomic.Bool
}

func New[T any]() *MichaelScottQueue[T] {
	dummy := &item[T]{}
	output := &MichaelScottQueue[T]{}
	output.head.Store(dummy)
	output.tail.Store(dummy)
	return output
}

func (slf *MichaelScottQueue[T]) Enqueue(value T) bool {
	node := &item[T]{value: value}
	for {
		if slf.closed.Load() {
			return false
		}

		tail := slf.tail.Load()
		if tail.next.CompareAndSwap(nil, node) {
			slf.tail.CompareAndSwap(tail, node)
			return true
		}

		slf.tail.CompareAndSwap(tail, tail.next.Load())
	}
}

func (slf *MichaelScottQueue[T]) Dequeue() (T, bool) {
	for {
		if slf.closed.Load() {
			return *new(T), false
		}

		head := slf.head.Load()
		next := head.next.Load()
		if next != nil {
			if slf.head.CompareAndSwap(head, next) {
				return next.value, true
			}
		}
	}
}

func (slf *MichaelScottQueue[T]) Close() { slf.closed.Store(true) }
