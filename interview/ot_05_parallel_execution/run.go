package ot05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run производит запуск массива задач на N независимых исполнителях.
// Если во время работы происходит суммарно M ошибок, то исполнители закончат текущую задачу и завершат работу.
func Run(tasks []Task, n, m int) error {
	errCounter := new(atomic.Int64)

	wg, ch := new(sync.WaitGroup), make(chan Task)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case task, ok := <-ch:
					switch {
					case !ok:
						return
					case errCounter.Load() >= int64(m):
						continue
					case task() != nil:
						errCounter.Add(1)
					}
				}
			}
		}()
	}

	for _, v := range tasks {
		ch <- v
	}
	close(ch)
	wg.Wait()

	if errCounter.Load() >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}

// RunWithSliceConcurrentAccess аналогичен Run с той разницей, что использует конкурентное чтение слайса задач.
// Допустимо, если гарантируется отсутствие записи в слайс во время исполнения.
func RunWithSliceConcurrentAccess(tasks []Task, n, m int) error {
	nextTask, errCounter, l := new(atomic.Int64), new(atomic.Int64), int64(len(tasks))

	wg := new(sync.WaitGroup)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := nextTask.Add(1); i <= l && errCounter.Load() < int64(m); i = nextTask.Add(1) {
				if err := tasks[i-1](); err != nil && errCounter.Add(1) >= int64(m) {
					return
				}
			}
		}()
	}
	wg.Wait()

	if errCounter.Load() >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}
