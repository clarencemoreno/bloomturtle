package bloomturtle

import (
	"testing"

	"github.com/clarencemoreno/bloomturtle/internal/event"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20, 30)
	if rl == nil {
		t.Error("NewRateLimiter returned nil")
	}
}

func TestNewStorekeeper(t *testing.T) {
	publisher := &event.BaseEventPublisher{} // Create a mock publisher
	sk := NewStorekeeper(publisher)
	if sk == nil {
		t.Error("NewStorekeeper returned nil")
	}
}
