package storekeeper

import (
	"context"
	"fmt"

	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
)

// Storekeeper struct
type Storekeeper struct {
	// Additional fields if necessary
}

// New creates a new Storekeeper
func New() *Storekeeper {
	return &Storekeeper{}
}

// Check method checks the key and returns a boolean
func (sk *Storekeeper) Check(key string) bool {
	// Implementation of key check
	return true // Placeholder
}

// HandleEvent handles events from the EventPublisher
func (sk *Storekeeper) HandleEvent(ctx context.Context, event event.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		switch e := event.(type) {
		case ratelimiter.RateLimitEvent:
			fmt.Println("Storekeeper received event:", e.Message)
		default:
			return fmt.Errorf("unknown event type: %T", e)
		}
	}
	return nil
}
