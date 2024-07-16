package bloomturtle

import (
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
	"github.com/clarencemoreno/bloomturtle/internal/storekeeper"
)

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(size uint, hashFuncs []func([]byte) uint) *ratelimiter.RateLimiter {
	return ratelimiter.New(size, hashFuncs)
}

// NewStorekeeper creates a new Storekeeper
func NewStorekeeper() *storekeeper.Storekeeper {
	return storekeeper.New()
}
