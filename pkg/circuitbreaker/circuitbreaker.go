package circuitbreaker

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var ErrBlocked = errors.New("blocked")

// CircuitBreaker ведёт учёт ошибок, полученных из входящей функции, и при превышении лимита ошибок в единицу времени
// производит блокировку исполнения функций. Сначала жёсткая блокировка продолжается в течение заданного периода,
// и во время жёсткой блокировки попытка исполнения функции приведёт к ошибке ErrBlocked.
// Затем включается мягкая блокировка, в течение которой лимит устанавливается на 20% от максимума.
// Если в течение мягкой блокировки лимит ошибок не был превышен, блокировки снимаются.
type CircuitBreaker interface {
	Eval(func() (any, error)) (any, error)
}

type cb struct {
	errCounter               atomic.Int64
	signal                   chan struct{}
	limit                    int64
	checkPeriod, blockPeriod time.Duration
}

func New(ctx context.Context, limit int64, checkPeriod, blockPeriod time.Duration) CircuitBreaker {
	cb := &cb{signal: make(chan struct{}), limit: limit, checkPeriod: checkPeriod, blockPeriod: blockPeriod}
	go cb.watchdog(ctx)
	return cb
}

func (slf *cb) Eval(f func() (any, error)) (any, error) {
	if slf.errCounter.Load() >= slf.limit {
		return nil, ErrBlocked
	}

	result, err := f()
	if err != nil {
		if slf.errCounter.Add(1) == slf.limit {
			select {
			case slf.signal <- struct{}{}:
			default:
			}

		}
	}
	return result, err
}

func (slf *cb) watchdog(ctx context.Context) {
	ticker := time.NewTicker(slf.checkPeriod)
	defer ticker.Stop()

main:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slf.errCounter.Store(0)
		case <-slf.signal:
			slf.errCounter.Store(slf.limit)
		internal:
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(slf.blockPeriod):
				}

				slf.errCounter.Store(slf.limit - slf.limit/5)

				select {
				case <-ctx.Done():
					return
				case <-slf.signal:
					continue internal
				case <-time.After(slf.blockPeriod):
					continue main
				}
			}
		}
	}
}
