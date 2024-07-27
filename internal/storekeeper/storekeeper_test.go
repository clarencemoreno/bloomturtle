package storekeeper

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter_bloom"
)

// TestStorekeeper_Check tests the Storekeeper with the actual BaseEventPublisher.
func TestStorekeeper_Check(t *testing.T) {
	// Create and start a real BaseEventPublisher
	basePublisher := event.NewBaseEventPublisher()
	basePublisher.Start()
	defer func() {
		// Ensure the publisher shuts down gracefully
		if err := basePublisher.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown publisher: %v", err)
		}
	}()

	// Create Storekeeper
	sk := New()

	// Register Storekeeper as a listener to the base publisher
	basePublisher.AddListener(sk)

	// Check function before publishing the event
	if sk.Check("key") {
		t.Errorf("expected Check to return false before publishing the event")
	}

	// Create a rate limit event
	limitEvent := ratelimiter_bloom.RateLimitEvent{
		Key:                 "key",
		Message:             "Rate limit exceeded",
		Timestamp:           time.Now(),
		ExpirationTimestamp: time.Now().Add(10 * time.Millisecond),
	}

	// Publish the event
	err := basePublisher.PublishEvent(limitEvent)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Wait for event processing
	time.Sleep(2 * time.Millisecond) // Adjust if needed based on event processing time

	// Check function after publishing the event
	if !sk.Check("key") {
		t.Errorf("expected Check to return true after publishing the event")
	}

	// Wait for the expiration of the event
	time.Sleep(10 * time.Millisecond)

	// Check function after the event has expired
	if sk.Check("key") {
		t.Errorf("expected Check to return false after the event has expired")
	}
}

func TestPerformanceStorekeeper(t *testing.T) {
	last := time.Now()
	// Create Storekeeper
	sk := New()

	for i := 0; i < 10; i++ {
		sk.Check("key")
		cur := time.Now()
		fmt.Println("last", cur.Sub(last))
		last = cur
	}
}

// TestStorekeeper_ConcurrentEvents tests the Storekeeper under concurrent event publishing for both different and same keys.
func TestStorekeeper_ConcurrentEvents(t *testing.T) {
	// Create and start a real BaseEventPublisher
	basePublisher := event.NewBaseEventPublisher()
	basePublisher.Start()
	defer func() {
		// Ensure the publisher shuts down gracefully
		if err := basePublisher.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to shutdown publisher: %v", err)
		}
	}()

	// Create Storekeeper
	sk := New()

	// Register Storekeeper as a listener to the base publisher
	basePublisher.AddListener(sk)

	// Number of concurrent events
	const numEvents = 1000
	const numKeys = 10

	var wg sync.WaitGroup
	wg.Add(numEvents * numKeys * 2)

	// Publish multiple events concurrently for different keys
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key-diff-%d", i)
		for j := 0; j < numEvents; j++ {
			go func(key string) {
				defer wg.Done()
				limitEvent := ratelimiter_bloom.RateLimitEvent{
					Key:                 key,
					Message:             "Rate limit exceeded",
					Timestamp:           time.Now(),
					ExpirationTimestamp: time.Now().Add(100 * time.Second),
				}
				err := basePublisher.PublishEvent(limitEvent)
				if err != nil {
					t.Errorf("failed to publish event: %v", err)
				}
			}(key)
		}
	}

	// Publish multiple events concurrently for the same keys
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key-same-%d", i)
		for j := 0; j < numEvents; j++ {
			go func(key string) {
				defer wg.Done()
				limitEvent := ratelimiter_bloom.RateLimitEvent{
					Key:                 key,
					Message:             "Rate limit exceeded",
					Timestamp:           time.Now(),
					ExpirationTimestamp: time.Now().Add(1000 * time.Second),
				}
				err := basePublisher.PublishEvent(limitEvent)
				if err != nil {
					t.Errorf("failed to publish event: %v", err)
				}
			}(key)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	time.Sleep(2 * time.Second)

	// Verify the Storekeeper's state for different keys
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key-diff-%d", i)
		if !sk.Check(key) {
			t.Errorf("expected Check to return true for key: %s", key)
		}
	}

	// Verify the Storekeeper's state for same keys
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key-same-%d", i)
		if !sk.Check(key) {
			t.Errorf("expected Check to return true for key: %s", key)
		}
	}

	// Wait for the expiration of the events
	// time.Sleep(2 * time.Second)

	// // Verify the Storekeeper's state after expiration for different keys
	// for i := 0; i < numKeys; i++ {
	// 	key := fmt.Sprintf("key-diff-%d", i)
	// 	if sk.Check(key) {
	// 		t.Errorf("expected Check to return false for key: %s after expiration", key)
	// 	}
	// }

	// // Verify the Storekeeper's state after expiration for same keys
	// for i := 0; i < numKeys; i++ {
	// 	key := fmt.Sprintf("key-same-%d", i)
	// 	if sk.Check(key) {
	// 		t.Errorf("expected Check to return false for key: %s after expiration", key)
	// 	}
	// }
}
