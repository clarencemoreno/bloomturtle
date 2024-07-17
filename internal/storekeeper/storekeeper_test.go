package storekeeper

import (
	"context"
	"testing"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
)

// mockEventPublisher is a mock implementation of EventPublisher for testing purposes.
type mockEventPublisher struct {
	eventChan chan event.Event
}

// NewMockEventPublisher creates a new mock EventPublisher.
func NewMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		eventChan: make(chan event.Event, 10),
	}
}

func (m *mockEventPublisher) Start() {
	go m.run()
}

func (m *mockEventPublisher) PublishEvent(event event.Event) error {
	m.eventChan <- event
	return nil
}

func (m *mockEventPublisher) AddListener(listener event.EventListener) {
	// Implement if needed for testing
}

func (m *mockEventPublisher) Shutdown(ctx context.Context) error {
	close(m.eventChan)
	return nil
}

func (m *mockEventPublisher) run() {
	for range m.eventChan {
		// Simulate event processing
	}
}

func TestStorekeeper_Check(t *testing.T) {
	// Create a mock event publisher
	mockPublisher := NewMockEventPublisher()
	defer mockPublisher.Shutdown(context.Background())

	// Create Storekeeper with mock event publisher
	sk := New(mockPublisher)

	// Simulate adding an item to the cache
	event := ratelimiter.RateLimitEvent{
		Key:                 "key",
		Message:             "Rate limit reached",
		Timestamp:           time.Now(),
		ExpirationTimestamp: time.Now().Add(1 * time.Hour), // Set expiration to 1 hour in the future
	}

	// Add event to the Storekeeper cache
	sk.handleRateLimitEvent(event)

	// Test Check function
	if !sk.Check("key") {
		t.Errorf("expected Check to return true")
	}
}
