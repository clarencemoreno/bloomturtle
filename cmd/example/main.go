package main

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencemoreno/bloomturtle"
	"github.com/clarencemoreno/bloomturtle/internal/event"
)

func main() {
	// Define the RateLimiter parameters
	primaryCapacity := 10
	secondaryCapacity := 5
	rate := 1 // Tokens per second

	// Create the RateLimiter
	rl := bloomturtle.NewRateLimiter(primaryCapacity, secondaryCapacity, rate)
	defer rl.Shutdown(context.Background())

	// Create the Event Publisher
	eventPublisher := event.NewBaseEventPublisher()
	defer eventPublisher.Shutdown(context.Background())

	// Create the Storekeeper and add it as a listener
	sk := bloomturtle.NewStorekeeper(eventPublisher)
	rl.AddListener(sk)

	// Simulate requests and print results on the same line
	for i := 0; i < 150; i++ {
		key := fmt.Sprintf("key-%d", i)
		if rl.Allow(key) {
			fmt.Printf("\rRequest allowed: %s", key)
		} else {
			fmt.Printf("\rRequest denied: %s", key)
		}
		time.Sleep(100 * time.Millisecond) // Sleep for visibility
	}

	// Move to the next line after loop is done
	fmt.Println()
}
