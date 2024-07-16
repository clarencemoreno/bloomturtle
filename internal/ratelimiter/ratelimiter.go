package ratelimiter

import (
	"context"
	"sync"

	"github.com/clarencemoreno/bloomturtle/internal/event"
)

// RateLimiter is a basic rate limiter implementation.
type RateLimiter struct {
	mu        sync.Mutex
	bitmap    []bool
	hashFuncs []func([]byte) uint
	eventPub  *event.BaseEventPublisher
	capacity  int
	count     int
	threshold int
}

// New creates a new RateLimiter.
func New(capacity uint, hashFuncs []func([]byte) uint) *RateLimiter {
	rl := &RateLimiter{
		bitmap:    make([]bool, capacity),
		hashFuncs: hashFuncs,
		capacity:  int(capacity), // Convert uint to int
		eventPub:  event.NewBaseEventPublisher(),
		threshold: int(capacity), // Convert uint to int
	}
	rl.eventPub.Start()
	return rl
}

// Add adds data to the rate limiter.
func (rl *RateLimiter) Add(data []byte) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, hashFunc := range rl.hashFuncs {
		index := hashFunc(data) % uint(len(rl.bitmap))
		rl.bitmap[index] = true
	}

	rl.count++
	if rl.count >= rl.threshold {
		rl.eventPub.PublishEvent(RateLimitEvent{Message: "Rate limit reached"})
	}

	return nil
}

// Contains checks if data is contained in the rate limiter.
func (rl *RateLimiter) Contains(data []byte) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, hashFunc := range rl.hashFuncs {
		index := hashFunc(data) % uint(len(rl.bitmap))
		// Check if the data is present at the index
		if rl.bitmap[index] {
			return true, nil // Data is present, return true
		}
	}

	// If none of the hash functions indicate the data is present, return false
	return false, nil
}

// AddListener adds an event listener to the rate limiter.
func (rl *RateLimiter) AddListener(listener event.EventListener) {
	rl.eventPub.AddListener(listener)
}

// Shutdown stops the event publisher and waits for it to finish.
func (rl *RateLimiter) Shutdown(ctx context.Context) error {
	return rl.eventPub.Shutdown(ctx)
}

// RateLimitEvent represents an event when the rate limit is reached.
type RateLimitEvent struct {
	Message string // Fixed the syntax error here
}
