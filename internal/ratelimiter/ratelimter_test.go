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

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(10, 5, 1)
	defer rl.Shutdown(context.Background())

	listener := &mockRateLimitListener{}
	rl.AddListener(listener)

	// Use up all primary and secondary tokens
	for i := 0; i < 15; i++ {
		if !rl.Allow("key") {
			t.Fatalf("expected request %d to be allowed", i)
		}
	}

	// The next request should be rejected since both buckets are exhausted
	if rl.Allow("key") {
		t.Fatalf("expected request to be rejected")
	}

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	listener.mu.Lock()
	defer listener.mu.Unlock()
	if len(listener.receivedEvents) != 1 {
		t.Errorf("expected 1 event to be received, got %d", len(listener.receivedEvents))
	}
	if event, ok := listener.receivedEvents[0].(RateLimitEvent); !ok || event.Message != "Both primary and secondary buckets are exhausted." {
		t.Errorf("expected RateLimitEvent with message 'Both primary and secondary buckets are exhausted.', got %v", listener.receivedEvents[0])
	}
}
