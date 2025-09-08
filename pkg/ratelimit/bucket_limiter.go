package ratelimit

import (
	"sync"
	"time"
)

type LeakyBucketLimiter struct {
	bucket, stop chan struct{}
	interval     time.Duration
	stopOnce     sync.Once
}

func NewLeakyBucketLimiter(limit int, window time.Duration) *LeakyBucketLimiter {
	limiter := &LeakyBucketLimiter{
		bucket:   make(chan struct{}, limit),
		stop:     make(chan struct{}),
		interval: window / time.Duration(limit),
	}

	go limiter.run()

	return limiter
}

func (slf *LeakyBucketLimiter) Allow() bool {
	select {
	case slf.bucket <- struct{}{}:
		return true
	default:
		return false
	}
}

func (slf *LeakyBucketLimiter) Stop() { slf.stopOnce.Do(func() { close(slf.stop) }) }

func (slf *LeakyBucketLimiter) run() {
	ticker := time.NewTicker(slf.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case <-slf.bucket:
			default:
			}
		case <-slf.stop:
			return
		}
	}
}
