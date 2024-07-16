package main

import (
	"context"
	"fmt"

	"github.com/clarencemoreno/bloomturtle"
)

func main() {
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint { return uint(data[0]) },
		func(data []byte) uint { return uint(data[1]) },
	}

	rl := bloomturtle.NewRateLimiter(1000, hashFuncs)
	sk := bloomturtle.NewStorekeeper()

	// Register storekeeper as an event listener
	rl.AddListener(sk)

	for i := 0; i < 100; i++ {
		if err := rl.Add([]byte(fmt.Sprintf("data%d", i))); err != nil {
			fmt.Printf("error adding data: %v\n", err)
		}
	}

	contains, err := rl.Contains([]byte("data0"))
	if err != nil {
		fmt.Printf("error checking data: %v\n", err)
	}
	fmt.Println(contains) // Output: true

	contains, err = rl.Contains([]byte("data100"))
	if err != nil {
		fmt.Printf("error checking data: %v\n", err)
	}
	fmt.Println(contains) // Output: false

	// Give some time for the asynchronous event handling to complete
	fmt.Println("Press Enter to shutdown...")
	fmt.Scanln()

	// Shutdown the event publisher gracefully
	if err := rl.Shutdown(context.Background()); err != nil {
		fmt.Printf("error shutting down: %v\n", err)
	}
}
