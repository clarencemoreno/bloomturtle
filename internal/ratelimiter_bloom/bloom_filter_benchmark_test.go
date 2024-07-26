package ratelimiter_bloom

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func BenchmarkCASHeavyWrite(b *testing.B) {
	size := uint32(100)
	initialCapacity := uint32(60000)
	bfc := NewBloomFilterCounter(size, initialCapacity)
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	var wg sync.WaitGroup
	numGoroutines := 10000 // Use a large number of goroutines

	// Benchmark loop
	for n := 0; n < b.N; n++ {
		wg.Add(numGoroutines * len(keys))
		done := make(chan struct{})

		// Start goroutines for incrementing and decrementing
		for _, key := range keys {
			for i := 0; i < numGoroutines; i++ {
				go func(k string) {
					defer wg.Done()
					for {
						select {
						case <-done:
							return
						default:
							bfc.Increment(k, 1)

							// Random sleep between 10ms and 50ms
							time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)

							bfc.DecrementWithRetryDecorator(bfc.Decrement)(k, 1)

							// Random sleep between 10ms and 50ms
							time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)
						}
					}
				}(key)
			}
		}

		// Run for a specific duration
		time.Sleep(5 * time.Second)

		// Signal all goroutines to stop
		close(done)

		// Wait for all goroutines to finish
		wg.Wait()
	}
}

func BenchmarkMutexHeavyWrite(b *testing.B) {
	size := uint32(100)
	initialCapacity := uint32(60000)
	bfc := NewBloomFilterCounter(size, initialCapacity)
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	var wg sync.WaitGroup
	numGoroutines := 10000 // Use a large number of goroutines

	// Benchmark loop
	for n := 0; n < b.N; n++ {
		wg.Add(numGoroutines * len(keys))
		done := make(chan struct{})

		// Start goroutines for incrementing and decrementing
		for _, key := range keys {
			for i := 0; i < numGoroutines; i++ {
				go func(k string) {
					defer wg.Done()
					for {
						select {
						case <-done:
							return
						default:
							bfc.Increment(k, 1)

							// Random sleep between 10ms and 50ms
							time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)

							bfc.Decrement(k, 1)

							// Random sleep between 10ms and 50ms
							time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)
						}
					}
				}(key)
			}
		}

		// Run for a specific duration
		time.Sleep(5 * time.Second)

		// Signal all goroutines to stop
		close(done)

		// Wait for all goroutines to finish
		wg.Wait()
	}
}
