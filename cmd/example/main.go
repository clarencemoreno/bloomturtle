package main

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter_bloom"
	"github.com/clarencemoreno/bloomturtle/internal/storekeeper"
)

func main() {
	// Define the RateLimiter parameters
	capacity := uint32(10)
	rate := 2

	// Create the RateLimiter
	rl := ratelimiter_bloom.NewRateLimiter(capacity, rate)
	defer rl.Shutdown(context.Background())

	// // Create the Event Publisher
	// eventPublisher := event.NewBaseEventPublisher()
	// eventPublisher.Start()
	// defer eventPublisher.Shutdown(context.Background())

	// // Create the Storekeeper and add it as a listener
	sk := storekeeper.New()
	// eventPublisher.AddListener(sk)
	rl.AddListener(sk)

	key := "testKey"

	// Test initial state with tokens full
	for i := 0; i < int(capacity); i++ {
		if !rl.Allow(key) {
			fmt.Printf("Expected request %d to be allowed, but it was denied\n", i+1)
		}
	}

	// Test exhaustion of tokens
	if rl.Allow(key) {
		fmt.Println("Expected request to be denied, but it was allowed")
	}

	time.Sleep(2 * time.Millisecond) // Wait for at least one token to be refilled
	println(sk.Check(key))
}
