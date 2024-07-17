package bloomturtle

import (
	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
	"github.com/clarencemoreno/bloomturtle/internal/storekeeper"
)

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(primaryCapacity int, secondaryCapacity int, rate int) *ratelimiter.RateLimiter {
	return ratelimiter.NewRateLimiter(primaryCapacity, secondaryCapacity, rate)
}

// NewStorekeeper creates a new Storekeeper with an EventPublisher
func NewStorekeeper(publisher event.EventPublisher) *storekeeper.Storekeeper {
	return storekeeper.New(publisher)
}
