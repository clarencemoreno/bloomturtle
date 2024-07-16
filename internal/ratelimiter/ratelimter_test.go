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

// func TestRateLimiter_AddContains(t *testing.T) {
// 	hashFuncs := []func([]byte) uint{
// 		func(data []byte) uint {
// 			if len(data) > 0 {
// 				return uint(data[0])
// 			}
// 			return 0
// 		},
// 	}

// 	rl := New(1000, hashFuncs)
// 	defer rl.Shutdown(context.Background())

// 	data := []byte("data0")
// 	err := rl.Add(data)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	contains, err := rl.Contains(data)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	if !contains {
// 		t.Errorf("expected data to be contained in the rate limiter")
// 	}

// 	// The test is expecting data100 to NOT be contained, so we don't add it
// 	// contains, err = rl.Contains([]byte("data100"))
// 	// if err != nil {
// 	// 	t.Fatalf("unexpected error: %v", err)
// 	// }
// 	// if contains {
// 	// 	t.Errorf("expected data not to be contained in the rate limiter")
// 	// }

// 	contains, err = rl.Contains([]byte("data100"))
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	if contains {
// 		t.Errorf("expected data100 not to be contained in the rate limiter")
// 	}
// }

func TestRateLimiter_EventPublishing(t *testing.T) {
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint {
			if len(data) > 0 {
				return uint(data[0])
			}
			return 0
		},
	}

	rl := New(10, hashFuncs)
	defer rl.Shutdown(context.Background())

	listener := &mockRateLimitListener{}
	rl.AddListener(listener)

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
	if len(listener.receivedEvents) != 1 {
		t.Errorf("expected 1 event to be received, got %d", len(listener.receivedEvents))
	}
	if event, ok := listener.receivedEvents[0].(RateLimitEvent); !ok || event.Message != "Rate limit reached" {
		t.Errorf("expected RateLimitEvent with message 'Rate limit reached', got %v", listener.receivedEvents[0])
	}
}
