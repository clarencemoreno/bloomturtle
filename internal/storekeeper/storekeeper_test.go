package storekeeper

import (
	"context"
	"testing"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
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
	limitEvent := ratelimiter.RateLimitEvent{
		Key:                 "key",
		Message:             "Rate limit exceeded",
		Timestamp:           time.Now(),
		ExpirationTimestamp: time.Now().Add(1 * time.Hour),
	}

	// Publish the event
	err := basePublisher.PublishEvent(limitEvent)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Wait for event processing
	time.Sleep(100 * time.Millisecond) // Adjust if needed based on event processing time

	// Check function after publishing the event
	if !sk.Check("key") {
		t.Errorf("expected Check to return true after publishing the event")
	}

}
