package limiter

import (
	"sync"
	"time"
)

type LeakyBucketRateLimiter struct {
	capacity      int
	ratePerSecond int
	tokens        int
	lastLeakTime  time.Time
	mu            sync.Mutex
}

func NewLeakyBucket(capacity, ratePerSecond int) *LeakyBucketRateLimiter {
	return &LeakyBucketRateLimiter{
		capacity:      capacity,
		ratePerSecond: ratePerSecond,
		tokens:        0,
		lastLeakTime:  time.Now(),
	}
}

// Allow проверяет, можно ли выполнить запрос
func (l *LeakyBucketRateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastLeakTime).Seconds()
	drainedTokens := int(elapsed * float64(l.ratePerSecond))

	if drainedTokens > 0 {
		l.tokens = max(0, l.tokens-drainedTokens)
		l.lastLeakTime = now
	}

	if l.tokens < l.capacity {
		l.tokens++
		return true
	}

	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
