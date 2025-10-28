package service

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type TokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate int // tokens per minute
	lastRefill time.Time
	mu         sync.Mutex
}

type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	config  RateLimitConfig
}

type RateLimitConfig struct {
	OrdersPerMin int
	WindowSize   time.Duration
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		config:  config,
	}
}

func (rl *RateLimiter) Allow(teamName string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket := rl.getBucket(teamName)
	bucket.refill()

	if bucket.tokens <= 0 {
		log.Warn().
			Str("teamName", teamName).
			Int("maxTokens", bucket.maxTokens).
			Msg("Rate limit exceeded")
		return false
	}

	bucket.tokens--
	return true
}

func (rl *RateLimiter) getBucket(teamName string) *TokenBucket {
	bucket := rl.buckets[teamName]
	if bucket != nil {
		return bucket
	}

	// Create new bucket
	bucket = &TokenBucket{
		tokens:     rl.config.OrdersPerMin,
		maxTokens:  rl.config.OrdersPerMin,
		refillRate: rl.config.OrdersPerMin,
		lastRefill: time.Now(),
	}
	rl.buckets[teamName] = bucket
	return bucket
}

func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	if elapsed < time.Minute {
		return
	}

	// Calculate tokens to add based on elapsed time
	minutes := elapsed.Minutes()
	tokensToAdd := int(minutes * float64(tb.refillRate))

	if tokensToAdd <= 0 {
		return
	}

	tb.tokens += tokensToAdd
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefill = now
}

func (rl *RateLimiter) GetStatus(teamName string) (current, max int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket := rl.buckets[teamName]
	if bucket == nil {
		return rl.config.OrdersPerMin, rl.config.OrdersPerMin
	}

	bucket.refill()
	return bucket.tokens, bucket.maxTokens
}
