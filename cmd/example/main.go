package main

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencemoreno/bloomturtle"
	"github.com/clarencemoreno/bloomturtle/internal/event"
	"github.com/clarencemoreno/bloomturtle/internal/ratelimiter"
)

func main() {
	// Create a mock event publisher
	mockPublisher := event.NewBaseEventPublisher()
	defer mockPublisher.Shutdown(context.Background())

	// Create RateLimiter with correct parameters
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint {
			if len(data) > 0 {
				return uint(data[0])
			}
			return 0
		},
		func(data []byte) uint {
			if len(data) > 1 {
				return uint(data[1])
			}
			return 0
		},
	}

	// Create the RateLimiter with appropriate parameters
	rl := bloomturtle.NewRateLimiter(1000, hashFuncs, 50)

	// Create Storekeeper with the mock event publisher
	sk := bloomturtle.NewStorekeeper(mockPublisher)

	// Add a listener to the EventPublisher to handle events
	mockPublisher.AddListener(sk)

	// Add data to the rate limiter
	data := []byte("data0")
	err := rl.Add(data)
	if err != nil {
		fmt.Printf("unexpected error: %v\n", err)
		return
	}

	// Check if the data is contained in the rate limiter
	contains, err := rl.Contains(data)
	if err != nil {
		fmt.Printf("unexpected error: %v\n", err)
		return
	}
	if !contains {
		fmt.Println("expected data0 to be contained in the rate limiter")
	}

	// Add more data to the rate limiter
	err = rl.Add([]byte("data100"))
	if err != nil {
		fmt.Printf("unexpected error: %v\n", err)
		return
	}

	contains, err = rl.Contains([]byte("data100"))
	if err != nil {
		fmt.Printf("unexpected error: %v\n", err)
		return
	}
	if !contains {
		fmt.Println("expected data100 to be contained in the rate limiter")
	}

	// Publish an event and handle it with Storekeeper
	event := ratelimiter.RateLimitEvent{
		Key:                 "testKey",
		Message:             "Rate limit reached",
		ExpirationTimestamp: time.Now().Add(time.Minute),
	}

	err = mockPublisher.PublishEvent(event)
	if err != nil {
		fmt.Printf("unexpected error publishing event: %v\n", err)
		return
	}

	// Simulate handling event
	err = sk.HandleEvent(context.Background(), event)
	if err != nil {
		fmt.Printf("unexpected error handling event: %v\n", err)
	}
}
