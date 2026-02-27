package bot

import (
	"sync"
	"testing"
	"time"
)

func TestNewContactRateLimiterValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewContactRateLimiter(0, time.Second); err == nil {
		t.Fatalf("expected invalid max error")
	}
	if _, err := NewContactRateLimiter(1, 0); err == nil {
		t.Fatalf("expected invalid window error")
	}
}

func TestContactRateLimiterAllowAndReset(t *testing.T) {
	t.Parallel()

	limiter, err := NewContactRateLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("new limiter: %v", err)
	}

	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }

	if !limiter.Allow(42) {
		t.Fatalf("expected first call to pass")
	}
	if !limiter.Allow(42) {
		t.Fatalf("expected second call to pass")
	}
	if limiter.Allow(42) {
		t.Fatalf("expected third call to be limited")
	}

	now = now.Add(time.Second)
	if !limiter.Allow(42) {
		t.Fatalf("expected limiter reset after window")
	}
}

func TestContactRateLimiterSeparateContacts(t *testing.T) {
	t.Parallel()

	limiter, err := NewContactRateLimiter(1, time.Minute)
	if err != nil {
		t.Fatalf("new limiter: %v", err)
	}

	if !limiter.Allow(1) {
		t.Fatalf("expected first contact to pass")
	}
	if !limiter.Allow(2) {
		t.Fatalf("expected second contact to pass")
	}
	if limiter.Allow(1) {
		t.Fatalf("expected repeated contact to be limited")
	}
}

func TestContactRateLimiterConcurrent(t *testing.T) {
	t.Parallel()

	limiter, err := NewContactRateLimiter(1000, time.Minute)
	if err != nil {
		t.Fatalf("new limiter: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			for j := 0; j < 16; j++ {
				_ = limiter.Allow(id)
			}
		}(int64(i % 4))
	}
	wg.Wait()
}
