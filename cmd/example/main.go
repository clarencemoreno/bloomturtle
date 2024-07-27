package main

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter_bloom"
)

// MockEventListener is a mock implementation of the EventListener interface for testing purposes.
type MockEventListener struct {
	receivedEvents []event.Event
}

func (mel *MockEventListener) HandleEvent(ctx context.Context, e event.Event) error {
	println("Handle Rate Limit exceeded")
	mel.receivedEvents = append(mel.receivedEvents, e)
	println("No. of events", len(mel.receivedEvents))
	return nil
}

func main() {
	capacity := uint32(10)
	rate := 2

	rl := ratelimiter_bloom.NewRateLimiter(capacity, rate)
	defer rl.Shutdown(context.Background())

	mockListener := &MockEventListener{}
	rl.AddListener(mockListener)

	key := "testKey"

	// Test initial state with tokens full
	for i := 0; i < int(capacity); i++ {
		if !rl.Allow(key) {
			println("Expected request %d to be allowed, but it was denied", i+1)
		}
	}

	// Test exhaustion of tokens
	if rl.Allow(key) {
		println("Expected request to be denied, but it was allowed")
	}
	println("here: no of events", len(mockListener.receivedEvents))
	// Check if the event was generated
	// Give it time to finish generating the event
	time.Sleep(10 * time.Millisecond) // Wait for at least one token to be refilled

	if len(mockListener.receivedEvents) != 1 {
		println("Expected 1 event to be generated, but got %d", len(mockListener.receivedEvents))
	}

	event := mockListener.receivedEvents[0].(ratelimiter_bloom.RateLimitEvent)

	if event.Key != key {
		println("Expected event key to be %s, but got %s", key, event.Key)
	}

	time.Sleep(1 * time.Second) // Wait for at least one token to be refilled

	if !rl.Allow(key) {
		println("Expected request to be allowed after refill, but it was denied")
	}
}
func main2() {
	// Define the RateLimiter parameters
	capacity := uint32(10)
	rate := 1 // Tokens per second

	// Create the RateLimiter
	rl := ratelimiter_bloom.NewRateLimiter(capacity, rate)

	defer rl.Shutdown(context.Background())

	// Create the Event Publisher
	eventPublisher := event.NewBaseEventPublisher()
	defer eventPublisher.Shutdown(context.Background())

	// Create the Storekeeper and add it as a listener
	sk := ratelimiter_bloom.NewStorekeeper(eventPublisher)
	rl.AddListener(sk)

	// Simulate requests and print results on the same line
	for i := uint32(0); i < capacity+1; i++ {
		key := fmt.Sprintf("key-%d", i)
		if rl.Allow(key) {
			// fmt.Printf("\rRequest allowed: %s", key)
			fmt.Printf("Request allowed: %s\n", key)
		} else {
			// fmt.Printf("\rRequest denied: %s", key)
			fmt.Printf("Request allowed: %s\n", key)
		}
		time.Sleep(100 * time.Millisecond) // Sleep for visibility
	}

	// Move to the next line after loop is done
	fmt.Println()
}
