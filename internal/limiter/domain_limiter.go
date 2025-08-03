package limiter

import (
	"sync"
)

type DomainLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*LeakyBucketRateLimiter
}

func NewDomainLimiter() *DomainLimiter {
	return &DomainLimiter{
		limiters: make(map[string]*LeakyBucketRateLimiter),
	}
}

func (dl *DomainLimiter) GetLimiter(domain string) *LeakyBucketRateLimiter {
	dl.mu.RLock()
	limiter, exists := dl.limiters[domain]
	dl.mu.RUnlock()

	if !exists {
		dl.mu.Lock()
		defer dl.mu.Unlock()
		// 1 запрос в секунду, можно накопить до 3
		limiter = NewLeakyBucket(3, 1)
		dl.limiters[domain] = limiter
	}

	return limiter
}
