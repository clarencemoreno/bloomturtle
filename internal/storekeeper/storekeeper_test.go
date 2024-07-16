package storekeeper

import (
	"context"
	"testing"

	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
)

func TestStorekeeper_Check(t *testing.T) {
	sk := New()
	if !sk.Check("key") {
		t.Errorf("expected Check to return true")
	}
}

func TestStorekeeper_HandleEvent(t *testing.T) {
	sk := New()

	event := ratelimiter.RateLimitEvent{Message: "Rate limit reached"}
	err := sk.HandleEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorekeeper_HandleUnknownEvent(t *testing.T) {
	sk := New()

	event := struct{ Message string }{Message: "Unknown event"}
	err := sk.HandleEvent(context.Background(), event)
	if err == nil {
		t.Errorf("expected error when handling unknown event")
	}
}
