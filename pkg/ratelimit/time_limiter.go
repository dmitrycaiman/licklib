package ratelimit

import (
	"sync"
	"time"
)

type TimeLimiter struct {
	list   []time.Time
	window time.Duration
	limit  int
	mu     sync.Mutex
}

func NewTimeLimiter(limit int64, window time.Duration) *TimeLimiter {
	return &TimeLimiter{
		list:   make([]time.Time, 0),
		window: window,
		limit:  int(limit),
	}
}

func (slf *TimeLimiter) Allow() bool {
	slf.mu.Lock()
	defer slf.mu.Unlock()

	now := time.Now()
	border := now.Add(-slf.window)

	for i, timestamp := range slf.list {
		if timestamp.After(border) {
			slf.list = slf.list[i:]
			break
		}
	}

	if len(slf.list) < slf.limit {
		slf.list = append(slf.list, now)
		return true
	}
	return false
}
