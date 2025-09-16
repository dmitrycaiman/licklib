package workerpool

import (
	"context"
	"sync"
)

// SingleFlight группирует одновременные запросы с одинаковыми входными данными.
// Если во время исполнения запроса поступили запросы с такими же аргументами,
// то они дождутся его исполнения и получат его результаты.
type SingleFlight[T1 comparable, T2 any] struct {
	storage map[T1]*request[T2]
	f       func(T1) (T2, error)
	mu      sync.Mutex
}

func NewSingleFlight[T1 comparable, T2 any](f func(T1) (T2, error)) *SingleFlight[T1, T2] {
	return &SingleFlight[T1, T2]{storage: map[T1]*request[T2]{}, f: f}
}

func (slf *SingleFlight[T1, T2]) Do(ctx context.Context, input T1) (T2, error) {
	slf.mu.Lock()
	if r, ok := slf.storage[input]; ok {
		slf.mu.Unlock()
		return r.wait(ctx)
	}

	r := &request[T2]{done: make(chan struct{})}
	slf.storage[input] = r
	slf.mu.Unlock()

	go func() {
		r.result, r.err = slf.f(input)

		slf.mu.Lock()
		close(r.done)
		delete(slf.storage, input)
		slf.mu.Unlock()
	}()

	return r.wait(ctx)
}

type request[T2 any] struct {
	result T2
	err    error
	done   chan struct{}
}

func (slf *request[T2]) wait(ctx context.Context) (T2, error) {
	select {
	case <-slf.done:
		return slf.result, slf.err
	case <-ctx.Done():
		return *new(T2), ctx.Err()
	}
}
