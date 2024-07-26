package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/clarencemoreno/bloomturtle"
	"github.com/clarencemoreno/bloomturtle/internal/event"
)

func TestMainFunction(t *testing.T) {
	// Define the RateLimiter parameters
	primaryCapacity := 10
	secondaryCapacity := 5
	rate := 1 // Tokens per second

	// Create the RateLimiter
	rl := bloomturtle.NewRateLimiter(primaryCapacity, secondaryCapacity, rate)
	defer rl.Shutdown(context.Background())

	// Create the Event Publisher
	eventPublisher := event.NewBaseEventPublisher()
	defer eventPublisher.Shutdown(context.Background())

	// Create the Storekeeper and add it as a listener
	sk := bloomturtle.NewStorekeeper(eventPublisher)
	rl.AddListener(sk)

	// Track whether any request was denied
	var denied bool

	// Simulate requests
	for i := 0; i < 150; i++ {
		key := fmt.Sprintf("key-%d", i)
		if allowed := rl.Allow(key); allowed {
			t.Logf("Request allowed: %s", key)
		} else {
			t.Logf("Request denied: %s", key)
			denied = true
			break
		}
		time.Sleep(100 * time.Millisecond) // Adding delay to simulate time passage
	}

	// Assert that at least one request was denied
	if !denied {
		t.Errorf("Expected at least one request to be denied")
	}
}
