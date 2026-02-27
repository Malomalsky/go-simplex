package bot

import (
	"fmt"
	"sync"
	"time"
)

type ContactRateLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	now    func() time.Time

	gcEvery int
	ops     int

	buckets map[int64]contactRateBucket
}

type contactRateBucket struct {
	count   int
	resetAt time.Time
}

func NewContactRateLimiter(max int, window time.Duration) (*ContactRateLimiter, error) {
	if max <= 0 {
		return nil, fmt.Errorf("max must be > 0")
	}
	if window <= 0 {
		return nil, fmt.Errorf("window must be > 0")
	}
	return &ContactRateLimiter{
		max:     max,
		window:  window,
		now:     time.Now,
		gcEvery: 512,
		buckets: make(map[int64]contactRateBucket),
	}, nil
}

func (l *ContactRateLimiter) Allow(contactID int64) bool {
	if l == nil {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if l.gcEvery > 0 {
		l.ops++
		if l.ops%l.gcEvery == 0 {
			for id, b := range l.buckets {
				if !now.Before(b.resetAt) {
					delete(l.buckets, id)
				}
			}
		}
	}

	b := l.buckets[contactID]
	if b.resetAt.IsZero() || !now.Before(b.resetAt) {
		b.count = 0
		b.resetAt = now.Add(l.window)
	}
	if b.count >= l.max {
		l.buckets[contactID] = b
		return false
	}
	b.count++
	l.buckets[contactID] = b
	return true
}
