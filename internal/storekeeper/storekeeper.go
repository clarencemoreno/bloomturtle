package storekeeper

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter_bloom"
)

// Storekeeper struct
type Storekeeper struct {
	violatorCache unsafe.Pointer // Unsafe pointer to the violator cache
	mu            sync.RWMutex   // RWMutex for concurrent access
}

// New creates a new Storekeeper
func New() *Storekeeper {
	sk := &Storekeeper{}
	// Initialize the violator cache with an empty map
	initialCache := make(map[string]ratelimiter_bloom.RateLimitEvent)
	sk.violatorCache = unsafe.Pointer(&initialCache)
	return sk
}

// Check method checks the key and returns a boolean
func (sk *Storekeeper) Check(key string) bool {
	sk.mu.RLock()         // Lock for reading
	defer sk.mu.RUnlock() // Unlock after reading

	// Load the current violator cache (no atomic operation)
	cachePtr := (*map[string]ratelimiter_bloom.RateLimitEvent)(atomic.LoadPointer(&sk.violatorCache))
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
	fmt.Printf("Inside HandleEvent, event type: %T\n", event)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		switch e := event.(type) {
		case ratelimiter_bloom.RateLimitEvent:
			println("ratelimiter_bloom.RateLimitEvent")
			sk.handleRateLimitEvent(e)
		default:
			return fmt.Errorf("unknown event type: %T", e)
		}
	}
	return nil
}

// handleRateLimitEvent handles RateLimitEvent and updates the cache atomically
func (sk *Storekeeper) handleRateLimitEvent(e ratelimiter_bloom.RateLimitEvent) {
	for i := 0; i < 3; i++ {
		// Atomically load the current violator cache
		oldCachePtr := atomic.LoadPointer(&sk.violatorCache)
		oldCache := *(*map[string]ratelimiter_bloom.RateLimitEvent)(oldCachePtr)
		newCache := make(map[string]ratelimiter_bloom.RateLimitEvent)

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
			fmt.Println("Storekeeper received event:", e.Message)
			return // Successful swap, exit method
		}
		// If swap failed, another goroutine might have updated the cache, retry
	}

	// Fallback to mutex lock if atomic swap fails after 3 tries
	sk.mu.Lock()
	defer sk.mu.Unlock()

	// Atomically load the current violator cache again within the lock
	oldCachePtr := atomic.LoadPointer(&sk.violatorCache)
	oldCache := *(*map[string]ratelimiter_bloom.RateLimitEvent)(oldCachePtr)
	newCache := make(map[string]ratelimiter_bloom.RateLimitEvent)

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
	atomic.StorePointer(&sk.violatorCache, newCachePtr)

	fmt.Println("Storekeeper received event:", e.Message)
}
