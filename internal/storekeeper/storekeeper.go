package storekeeper

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
)

// Storekeeper struct
type Storekeeper struct {
	violatorCache unsafe.Pointer // Unsafe pointer to the violator cache
}

// New creates a new Storekeeper
func New() *Storekeeper {
	sk := &Storekeeper{}
	// Initialize the violator cache with an empty map
	initialCache := make(map[string]ratelimiter.RateLimitEvent)
	sk.violatorCache = unsafe.Pointer(&initialCache)
	return sk
}

// Check method checks the key and returns a boolean
func (sk *Storekeeper) Check(key string) bool {
	// Load the current violator cache (no atomic operation)
	cachePtr := (*map[string]ratelimiter.RateLimitEvent)(atomic.LoadPointer(&sk.violatorCache))
	cache := *cachePtr
	value, ok := cache[key]
	if !ok {
		return false
	}
	// Check if the key's expiration timestamp has passed
	if time.Now().After(value.ExpirationTimestamp) {
		return false
	}
	return true
}

// HandleEvent handles events from the EventPublisher
func (sk *Storekeeper) HandleEvent(ctx context.Context, event event.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		switch e := event.(type) {
		case ratelimiter.RateLimitEvent:
			sk.handleRateLimitEvent(e)
		default:
			return fmt.Errorf("unknown event type: %T", e)
		}
	}
	return nil
}

// handleRateLimitEvent handles RateLimitEvent and updates the cache atomically
func (sk *Storekeeper) handleRateLimitEvent(e ratelimiter.RateLimitEvent) {
	for {
		// Atomically load the current violator cache
		oldCachePtr := atomic.LoadPointer(&sk.violatorCache)
		oldCache := *(*map[string]ratelimiter.RateLimitEvent)(oldCachePtr)
		newCache := make(map[string]ratelimiter.RateLimitEvent)

		// Copy old cache, ignoring expired items
		for k, v := range oldCache {
			if !time.Now().After(v.ExpirationTimestamp) {
				newCache[k] = v
			}
		}

		// Add the new event to the new cache
		newCache[e.Key] = e

		// Perform atomic swap
		newCachePtr := unsafe.Pointer(&newCache)
		if atomic.CompareAndSwapPointer(&sk.violatorCache, oldCachePtr, newCachePtr) {
			break // Successful swap, exit loop
		}
		// If swap failed, another goroutine might have updated the cache, retry
	}

	fmt.Println("Storekeeper received event:", e.Message)
}
