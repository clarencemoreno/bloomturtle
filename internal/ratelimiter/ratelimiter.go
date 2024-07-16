package ratelimiter

import (
	"context"
	"sync"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
)

// RateLimiter is a basic rate limiter implementation.
type RateLimiter struct {
	mu        sync.Mutex
	bitmap    []uint32
	hashFuncs []func([]byte) uint
	eventPub  *event.BaseEventPublisher
	capacity  int
	count     int
	threshold int
}

// New creates a new RateLimiter.
func New(capacity uint, hashFuncs []func([]byte) uint, threshold uint) *RateLimiter {
	rl := &RateLimiter{
		bitmap:    make([]uint32, capacity),
		hashFuncs: hashFuncs,
		capacity:  int(capacity),
		eventPub:  event.NewBaseEventPublisher(),
		threshold: int(threshold),
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
		rl.bitmap[index]++ // Increment the counter at the index
	}

	rl.count++
	if rl.count >= rl.threshold {
		err := rl.eventPub.PublishEvent(RateLimitEvent{
			Message: "Rate limit reached",
			Key:     string(data), // Assuming the data represents the key
			// Set an appropriate expiration timestamp if needed
			ExpirationTimestamp: time.Now().Add(10 * time.Minute), // Example timestamp
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Contains checks if data is contained in the rate limiter.
func (rl *RateLimiter) Contains(data []byte) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, hashFunc := range rl.hashFuncs {
		index := hashFunc(data) % uint(len(rl.bitmap))
		if rl.bitmap[index] > 0 {
			return true, nil // Data is likely present, return true
		}
	}

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

// RateLimitEvent struct for rate limit events.
type RateLimitEvent struct {
	Key                 string
	Message             string
	ExpirationTimestamp time.Time
}
