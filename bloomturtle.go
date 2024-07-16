package bloomturtle

import (
	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
	"github.com/clarencemoreno/bloomturtle/internal/storekeeper"
)

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(size uint, hashFuncs []func([]byte) uint, threshold uint) *ratelimiter.RateLimiter {
	return ratelimiter.New(size, hashFuncs, threshold)
}

// NewStorekeeper creates a new Storekeeper with an EventPublisher
func NewStorekeeper(publisher event.EventPublisher) *storekeeper.Storekeeper {
	return storekeeper.New(publisher)
}
