package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"
)

type QuotaLimiter struct {
	stop     chan struct{}
	counter  atomic.Int64
	limit    int64
	window   time.Duration
	stopOnce sync.Once
}

func NewQuotaLimiter(limit int64, window time.Duration) *QuotaLimiter {
	limiter := &QuotaLimiter{limit: limit, window: window}

	go limiter.run()

	return limiter
}

func (slf *QuotaLimiter) Allow() bool { return slf.counter.Add(1) <= slf.limit }

func (slf *QuotaLimiter) Stop() { slf.stopOnce.Do(func() { close(slf.stop) }) }

func (slf *QuotaLimiter) run() {
	ticker := time.NewTicker(slf.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slf.counter.Store(0)
		case <-slf.stop:
			return
		}
	}
}
