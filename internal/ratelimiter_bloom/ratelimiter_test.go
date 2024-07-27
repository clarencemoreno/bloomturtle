package ratelimiter_bloom

import (
	"context"
	"testing"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
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

func TestRateLimiter(t *testing.T) {
	capacity := uint32(5)
	rate := 2

	rl := NewRateLimiter(capacity, rate)

	mockListener := &MockEventListener{}
	rl.AddListener(mockListener)

	key := "testKey"

	// Test initial state with tokens full
	for i := 0; i < int(capacity); i++ {
		if !rl.Allow(key) {
			t.Errorf("Expected request %d to be allowed, but it was denied", i+1)
		}
	}

	// Test exhaustion of tokens
	if rl.Allow(key) {
		t.Errorf("Expected request to be denied, but it was allowed")
	}
	println("here: no of events", len(mockListener.receivedEvents))
	// Check if the event was generated
	// Give it time to finish generating the event
	time.Sleep(10 * time.Millisecond) // Wait for at least one token to be refilled

	if len(mockListener.receivedEvents) != 1 {
		t.Errorf("Expected 1 event to be generated, but got %d", len(mockListener.receivedEvents))
	}

	event := mockListener.receivedEvents[0].(RateLimitEvent)

	if event.Key != key {
		t.Errorf("Expected event key to be %s, but got %s", key, event.Key)
	}

	time.Sleep(1 * time.Second) // Wait for at least one token to be refilled

	if !rl.Allow(key) {
		t.Errorf("Expected request to be allowed after refill, but it was denied")
	}
}
