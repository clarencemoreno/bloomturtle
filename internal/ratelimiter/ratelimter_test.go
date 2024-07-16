package ratelimiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/event"
)

type mockRateLimitListener struct {
	receivedEvents []event.Event
	mu             sync.Mutex
}

func (ml *mockRateLimitListener) HandleEvent(ctx context.Context, event event.Event) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.receivedEvents = append(ml.receivedEvents, event)
	return nil
}

func TestRateLimiter_EventPublishing(t *testing.T) {
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint {
			if len(data) > 0 {
				return uint(data[0])
			}
			return 0
		},
	}

	rl := New(100, hashFuncs, 10)
	defer rl.Shutdown(context.Background())

	listener := &mockRateLimitListener{}
	rl.AddListener(listener)

	// Add 10 items to trigger the rate limit
	for i := 0; i < 10; i++ {
		err := rl.Add([]byte{byte(i)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	listener.mu.Lock()
	defer listener.mu.Unlock()

	// Check if the rate limit event was received
	if len(listener.receivedEvents) != 1 {
		t.Errorf("expected 1 event to be received, got %d", len(listener.receivedEvents))
	}

	// Check if the received event is a RateLimitEvent with the correct message
	if event, ok := listener.receivedEvents[0].(RateLimitEvent); !ok || event.Message != "Rate limit reached" {
		t.Errorf("expected RateLimitEvent with message 'Rate limit reached', got %v", listener.receivedEvents[0])
	}
}

func TestRateLimiter_Contains(t *testing.T) {
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint {
			if len(data) > 0 {
				return uint(data[0])
			}
			return 0
		},
	}

	rl := New(100, hashFuncs, 10)
	defer rl.Shutdown(context.Background())

	// Add some data
	rl.Add([]byte{1})
	rl.Add([]byte{2})

	// Check if data exists
	contains, err := rl.Contains([]byte{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains {
		t.Errorf("expected data to be present, got false")
	}

	// Check if non-existent data is not present
	contains, err = rl.Contains([]byte{3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains {
		t.Errorf("expected data to be absent, got true")
	}
}
