package ratelimiter_bloom

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

// no retry test
func TestConcurrentIncrement(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"
	var wg sync.WaitGroup
	numGoroutines := 100
	incrementCount := 10

	// Use IncrementWithRetry method
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bfc.Increment(key, incrementCount)
		}()
	}

	wg.Wait()

	// Compute the index for the given key
	index := bfc.hash(key) % uint64(bfc.size)
	// The expected count should be the initial capacity plus the total increments
	expectedCount := uint64(capacity) + uint64(numGoroutines*incrementCount)

	if bfc.array[index] != expectedCount {
		t.Errorf("Expected array[%d] to be %d, got %d", index, expectedCount, bfc.array[index])
	}
}

func TestConcurrentDecrement(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"

	// Initialize with initial count based on capacity
	initialCount := uint32(10) // Initial capacity is used

	var wg sync.WaitGroup
	numGoroutines := 100
	decrementCount := 1

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bfc.Decrement(key, decrementCount)
		}()
	}

	wg.Wait()

	// Compute the index for the given key
	// The expected count should be the initial capacity minus the total decrements
	expectedCount := int(capacity) + int(initialCount) - numGoroutines*decrementCount
	if expectedCount < 0 {
		expectedCount = 0 // Ensure the count does not drop below zero
	}

	if bfc.CheckCount(key) != expectedCount {
		t.Errorf("Expected CheckCount to return %d, got %d", expectedCount, bfc.CheckCount(key))
	}
}

func TestConcurrentMultipleKeys(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	incrementCount := 10
	numGoroutines := 100

	var wg sync.WaitGroup

	// Increment each key concurrently using IncrementWithRetry
	for _, key := range keys {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				bfc.Increment(k, incrementCount)
			}(key)
		}
	}

	wg.Wait()

	// Verify the counts for each key
	for _, key := range keys {
		index := bfc.hash(key) % uint64(bfc.size)
		expectedCount := uint64(capacity) + uint64(numGoroutines*incrementCount)

		if bfc.array[index] != expectedCount {
			t.Errorf("For key '%s': Expected array[%d] to be %d, got %d", key, index, expectedCount, bfc.array[index])
		}
	}

	// Verify that untouched slots have initial capacity
	expectedInitialCapacity := uint64(capacity)
	for i := 0; i < int(size); i++ {
		if bfc.array[i] != expectedInitialCapacity {
			// Check if this slot corresponds to any of the keys
			found := false
			for _, key := range keys {
				if bfc.hash(key)%uint64(bfc.size) == uint64(i) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Untouched slot %d should have initial capacity %d, got %d", i, capacity, bfc.array[i])
			}
		}
	}
}

// with retry tests
func TestConcurrentIncrementWithRetry(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"
	var wg sync.WaitGroup
	numGoroutines := 100
	incrementCount := 10

	// Use IncrementWithRetry method
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			incrementWithRetry := bfc.IncrementWithRetryDecorator(bfc.Increment)
			incrementWithRetry(key, incrementCount)
		}()
	}

	wg.Wait()

	// Compute the index for the given key
	index := bfc.hash(key) % uint64(bfc.size)
	// The expected count should be the initial capacity plus the total increments
	expectedCount := uint64(capacity) + uint64(numGoroutines*incrementCount)

	if bfc.array[index] != expectedCount {
		t.Errorf("Expected array[%d] to be %d, got %d", index, expectedCount, bfc.array[index])
	}
}

func TestConcurrentDecrementWithRetry(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"

	// Initialize with initial count based on capacity
	initialCount := uint32(10) // Initial capacity is used

	var wg sync.WaitGroup
	numGoroutines := 100
	decrementCount := 1

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decrementWithRetry := bfc.DecrementWithRetryDecorator(bfc.Increment)
			decrementWithRetry(key, decrementCount)
		}()
	}

	wg.Wait()

	// Compute the index for the given key
	// The expected count should be the initial capacity minus the total decrements
	expectedCount := int(capacity) + int(initialCount) - numGoroutines*decrementCount
	if expectedCount < 0 {
		expectedCount = 0 // Ensure the count does not drop below zero
	}

	if bfc.CheckCount(key) != expectedCount {
		t.Errorf("Expected CheckCount to return %d, got %d", expectedCount, bfc.CheckCount(key))
	}
}

func TestConcurrentIsEmpty(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decrementWithRetry := bfc.DecrementWithRetryDecorator(bfc.Increment)
			decrementWithRetry(key, 1)
		}()
	}

	wg.Wait()

	// Check if the key is empty after decrements
	if !bfc.IsEmpty(key) {
		t.Errorf("Expected IsEmpty to return true, but it returned false")
	}
}

func TestConcurrentMultipleKeysIncrementWithRetry(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	incrementCount := 10
	numGoroutines := 100

	var wg sync.WaitGroup

	// Increment each key concurrently using IncrementWithRetry

	for _, key := range keys {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				bfc.IncrementWithRetryDecorator(bfc.Increment)(k, incrementCount)

			}(key)
		}
	}

	wg.Wait()

	// Verify the counts for each key
	for _, key := range keys {
		index := bfc.hash(key) % uint64(bfc.size)
		expectedCount := uint64(capacity) + uint64(numGoroutines*incrementCount)

		if bfc.array[index] != expectedCount {
			t.Errorf("For key '%s': Expected array[%d] to be %d, got %d", key, index, expectedCount, bfc.array[index])
		}
	}

	// Verify that untouched slots have initial capacity
	expectedInitialCapacity := uint64(capacity)
	for i := 0; i < int(size); i++ {
		if bfc.array[i] != expectedInitialCapacity {
			// Check if this slot corresponds to any of the keys
			found := false
			for _, key := range keys {
				if bfc.hash(key)%uint64(bfc.size) == uint64(i) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Untouched slot %d should have initial capacity %d, got %d", i, capacity, bfc.array[i])
			}
		}
	}
}

func TestConcurrentIntermingledIncrementDecrement(t *testing.T) {
	size := uint32(100)
	initialCapacity := uint32(60000)
	bfc := NewBloomFilterCounter(size, initialCapacity)
	key := "test_key"
	var wg sync.WaitGroup
	numGoroutines := 10000 // Use a large number of goroutines

	// Create a channel to signal when all goroutines are done
	done := make(chan struct{})

	// Start goroutines for incrementing and decrementing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					// Intermingled operations:
					// - Increment by 1
					// - Decrement by 1
					bfc.IncrementWithRetryDecorator(bfc.Increment)(key, 1)

					// Random sleep between 10ms and 50ms
					time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)
					bfc.DecrementWithRetryDecorator(bfc.Decrement)(key, 1)

					// Random sleep between 10ms and 50ms
					time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)
				}
			}
		}()
	}

	// Wait for a short period to allow goroutines to run
	// time.Sleep(2 * time.Second)

	// Signal all goroutines to stop
	close(done)

	// Wait for all goroutines to finish
	wg.Wait()

	// The expected count should be back to the initial capacity
	expectedCount := int(initialCapacity)

	// Check the actual count
	actualCount := bfc.CheckCount(key)
	if actualCount != expectedCount {
		t.Errorf("Expected count to be %d, got %d", expectedCount, actualCount)
	}
}

func TestConcurrentIntermingledIncrementDecrementMultipleKeys(t *testing.T) {
	size := uint32(100)
	initialCapacity := uint32(60000)
	bfc := NewBloomFilterCounter(size, initialCapacity)
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	var wg sync.WaitGroup
	numGoroutines := 10000 // Use a large number of goroutines

	// Create a channel to signal when all goroutines are done
	done := make(chan struct{})

	// Start goroutines for incrementing and decrementing
	for _, key := range keys {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						// Intermingled operations:
						// - Increment by 1
						// - Decrement by 1
						bfc.IncrementWithRetryDecorator(bfc.Increment)(k, 1)

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

	// Signal all goroutines to stop
	close(done)

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify the counts for each key
	for _, key := range keys {
		expectedCount := int(initialCapacity) // Should be back to the initial capacity

		// Check the actual count
		actualCount := bfc.CheckCount(key)
		if actualCount != expectedCount {
			t.Errorf("For key '%s': Expected count to be %d, got %d", key, expectedCount, actualCount)
		}
	}

	// Verify that untouched slots have initial capacity
	expectedInitialCapacity := uint64(initialCapacity)
	for i := 0; i < int(size); i++ {
		if bfc.array[i] != expectedInitialCapacity {
			// Check if this slot corresponds to any of the keys
			found := false
			for _, key := range keys {
				if bfc.hash(key)%uint64(bfc.size) == uint64(i) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Untouched slot %d should have initial capacity %d, got %d", i, initialCapacity, bfc.array[i])
			}
		}
	}
}

// corner case to zero
func TestConcurrentIntermingledIncrementDecrementCloseToEmpty(t *testing.T) {
	size := uint32(100)
	initialCapacity := uint32(3)
	bfc := NewBloomFilterCounter(size, initialCapacity)
	key := "test_key"
	var wg sync.WaitGroup
	numGoroutines := 4 // Use a large number of goroutines

	// Create a channel to signal when all goroutines are done
	done := make(chan struct{})

	// Use sync.Once to ensure increment only runs once
	var once sync.Once

	// Start goroutines for incrementing and decrementing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					// Intermingled operations:
					// - Decrement by 1
					bfc.DecrementWithRetryDecorator(bfc.Decrement)(key, 1)
					// bfc.Decrement(key, 1)

					// Random sleep between 10ms and 50ms
					time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)

					// Increment only once using sync.Once
					once.Do(func() {
						// Increment the counter
						bfc.IncrementWithRetryDecorator(bfc.Increment)(key, 1)
					})
				}
			}
		}()
	}
	time.Sleep(10 * time.Millisecond)

	// Signal all goroutines to stop
	close(done)

	// Wait for all goroutines to finish
	wg.Wait()

	// The expected count should be either 0 or 1
	actualCount := bfc.CheckCount(key)
	if actualCount != 0 && actualCount != 1 {
		t.Errorf("Expected count to be 0 or 1, got %d", actualCount)
	}
}

func TestMultipleConcurrentIntermingledIncrementDecrementCloseToEmpty(t *testing.T) {
	size := uint32(100)
	initialCapacity := uint32(3)
	key := "test_key"

	for iteration := 0; iteration < 5; iteration++ {
		bfc := NewBloomFilterCounter(size, initialCapacity)
		var wg sync.WaitGroup

		// Randomly select the number of goroutines between 3 and 100
		rand.Seed(time.Now().UnixNano())
		numGoroutines := rand.Intn(98) + 3 // 98 + 3 = 101, so range is 3 to 100

		// Create a channel to signal when all goroutines are done
		done := make(chan struct{})

		// Use sync.Once to ensure increment only runs once
		var once sync.Once

		// Start goroutines for incrementing and decrementing
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						// Intermingled operations:
						// - Decrement by 1
						bfc.DecrementWithRetryDecorator(bfc.Decrement)(key, 1)

						// Random sleep between 10ms and 50ms
						time.Sleep(time.Duration(rand.Intn(41)+10) * time.Millisecond)

						// Increment only once using sync.Once
						once.Do(func() {
							// Increment the counter
							bfc.IncrementWithRetryDecorator(bfc.Increment)(key, 1)

						})
					}
				}
			}()
		}
		time.Sleep(10 * time.Millisecond)

		// Signal all goroutines to stop
		close(done)

		// Wait for all goroutines to finish
		wg.Wait()

		// The expected count should be either 0 or 1
		actualCount := bfc.CheckCount(key)
		if actualCount != 0 && actualCount != 1 {
			t.Errorf("Iteration %d: Expected count to be 0 or 1, got %d", iteration, actualCount)
		}
	}
}
