package ratelimiter_bloom

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
)

// Storekeeper is an implementation of EventListener that handles rate limit events.
type Storekeeper struct {
	eventPublisher *event.BaseEventPublisher
}

// NewStorekeeper creates a new Storekeeper.
func NewStorekeeper(eventPublisher *event.BaseEventPublisher) *Storekeeper {
	return &Storekeeper{eventPublisher: eventPublisher}
}

// HandleEvent processes rate limit events.
func (sk *Storekeeper) HandleEvent(ctx context.Context, e event.Event) error {
	event := e.(RateLimitEvent)
	fmt.Printf("\nRate limit event: Key=%s, Message=%s, Timestamp=%s\n", event.Key, event.Message, event.Timestamp)
	return nil
}

// RateLimiter controls the rate of requests using a leaky bucket algorithm
type RateLimiter struct {
	capacity   uint32              // Capacity of the bucket
	rate       int                 // Refill rate of the bucket (tokens per second)
	tokens     *BloomFilterCounter // Current number of tokens in the bucket
	lastRefill sync.Map            // Last time the bucket was refilled for each key
	eventPub   *event.BaseEventPublisher
	mutexes    sync.Map // Map of mutexes for each key
}

// NewRateLimiter creates a new RateLimiter with the specified capacity and refill rate
func NewRateLimiter(capacity uint32, rate int) *RateLimiter {
	rl := &RateLimiter{
		capacity: capacity,
		rate:     rate,
		tokens:   NewBloomFilterCounter(1000, capacity), // 1000 buckets, each with the specified capacity
		eventPub: event.NewBaseEventPublisher(),
	}
	rl.eventPub.Start()
	return rl
}

// getMutex retrieves or creates a mutex for the given key
// func (rl *RateLimiter) getMutex(key string) *sync.Mutex {
// 	mutex, _ := rl.mutexes.LoadOrStore(key, &sync.Mutex{})
// 	return mutex.(*sync.Mutex)
// }

// refill refills the bucket for the given key
func (rl *RateLimiter) refill(key string) {
	now := time.Now()
	lastRefillTime, _ := rl.lastRefill.LoadOrStore(key, now)
	elapsed := now.Sub(lastRefillTime.(time.Time)).Seconds()
	newTokens := int(elapsed * float64(rl.rate))
	if newTokens > 0 {
		rl.lastRefill.Store(key, now)
		rl.tokens.IncrementWithRetry(key, newTokens)
		if rl.tokens.CheckCount(key) > int(rl.capacity) {
			rl.tokens.DecrementWithRetry(key, rl.tokens.CheckCount(key)-int(rl.capacity))
		}
	}
}

// Allow checks if a request can be processed
func (rl *RateLimiter) Allow(key string) bool {
	rl.refill(key)

	if !rl.tokens.IsEmpty(key) {
		rl.tokens.DecrementWithRetry(key, 1)
		return true
	}

	// Generate event when the bucket is empty
	rl.generateEvent(key)
	println("Rate Limit Exceeded, ", rl.tokens.CheckCount(key))
	return false
}

// generateEvent is called when the bucket is exhausted
func (rl *RateLimiter) generateEvent(key string) {
	event := RateLimitEvent{
		Key:                 key,
		Timestamp:           time.Now(),
		Message:             "Bucket is exhausted.",
		ExpirationTimestamp: time.Now().Add(500 * time.Millisecond), // Expiration timestamp set to 0.5 seconds from now
	}
	rl.eventPub.PublishEvent(event)
	println("event published")
}

// AddListener adds an event listener to the rate limiter.
func (rl *RateLimiter) AddListener(listener event.EventListener) {
	rl.eventPub.AddListener(listener)
}

// Shutdown stops the event publisher and waits for it to finish.
func (rl *RateLimiter) Shutdown(ctx context.Context) error {
	return rl.eventPub.Shutdown(ctx)
}

// RateLimitEvent represents an event when the bucket is exhausted.
type RateLimitEvent struct {
	Key                 string
	Timestamp           time.Time
	Message             string
	ExpirationTimestamp time.Time
}
