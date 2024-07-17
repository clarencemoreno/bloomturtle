package ratelimiter

import (
	"context"
	"sync"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
)

// RateLimiter controls the rate of requests using a two-bucket leaky bucket algorithm
type RateLimiter struct {
	primaryCapacity   int       // Capacity of the primary bucket
	secondaryCapacity int       // Capacity of the secondary bucket (leaky bucket)
	rate              int       // Refill rate of the secondary bucket (tokens per second)
	primaryTokens     int       // Current number of tokens in the primary bucket
	secondaryTokens   int       // Current number of tokens in the secondary bucket
	lastRefillTime    time.Time // Last time the buckets were refilled
	eventPub          *event.BaseEventPublisher
	mutex             sync.Mutex // Mutex to synchronize access to the rate limiter
}

// NewRateLimiter creates a new RateLimiter with the specified capacities and refill rate
func NewRateLimiter(primaryCapacity int, secondaryCapacity int, rate int) *RateLimiter {
	rl := &RateLimiter{
		primaryCapacity:   primaryCapacity,
		secondaryCapacity: secondaryCapacity,
		rate:              rate,
		primaryTokens:     primaryCapacity,   // Initially, the primary bucket is full
		secondaryTokens:   secondaryCapacity, // Initially, the secondary bucket is full
		lastRefillTime:    time.Now(),
		eventPub:          event.NewBaseEventPublisher(),
	}
	rl.eventPub.Start()
	return rl
}

// refill refills the secondary bucket first, then the primary bucket if the secondary is full
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()
	newTokens := int(elapsed * float64(rl.rate))
	if newTokens > 0 {
		rl.secondaryTokens += newTokens
		if rl.secondaryTokens > rl.secondaryCapacity {
			excessTokens := rl.secondaryTokens - rl.secondaryCapacity
			rl.secondaryTokens = rl.secondaryCapacity
			rl.primaryTokens += excessTokens
			if rl.primaryTokens > rl.primaryCapacity {
				rl.primaryTokens = rl.primaryCapacity
			}
		}
		rl.lastRefillTime = now
	}
}

// Allow checks if a request can be processed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.refill()

	if rl.primaryTokens > 0 {
		rl.primaryTokens--
		return true
	} else if rl.secondaryTokens > 0 {
		rl.secondaryTokens--
		return true
	}

	rl.generateEvent(key)
	return false
}

// generateEvent is called when both buckets are exhausted
func (rl *RateLimiter) generateEvent(key string) {
	event := RateLimitEvent{
		Key:                 key,
		Timestamp:           time.Now(),
		Message:             "Both primary and secondary buckets are exhausted.",
		ExpirationTimestamp: time.Now().Add(500 * time.Millisecond), // Expiration timestamp set to 0.5 seconds from now
	}
	rl.eventPub.PublishEvent(event)
}

// AddListener adds an event listener to the rate limiter.
func (rl *RateLimiter) AddListener(listener event.EventListener) {
	rl.eventPub.AddListener(listener)
}

// Shutdown stops the event publisher and waits for it to finish.
func (rl *RateLimiter) Shutdown(ctx context.Context) error {
	return rl.eventPub.Shutdown(ctx)
}

// RateLimitEvent represents an event when both buckets are exhausted.
type RateLimitEvent struct {
	Key                 string
	Timestamp           time.Time
	Message             string
	ExpirationTimestamp time.Time
}
