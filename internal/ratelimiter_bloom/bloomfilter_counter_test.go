package ratelimiter_bloom

import (
	"sync"
	"testing"
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
	expectedCount := uint32(capacity) + uint32(numGoroutines*incrementCount)

	if bfc.bucket[index].array[0] != expectedCount {
		t.Errorf("Expected array[%d] to be %d, got %d", index, expectedCount, bfc.bucket[index].array[0])
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
		expectedCount := uint32(capacity) + uint32(numGoroutines*incrementCount)

		if bfc.bucket[index].array[0] != expectedCount {
			t.Errorf("For key '%s': Expected array[%d] to be %d, got %d", key, index, expectedCount, bfc.bucket[index].array[0])
		}
	}

	// Verify that untouched slots have initial capacity
	expectedInitialCapacity := uint32(capacity)
	for i := 0; i < int(size); i++ {
		if bfc.bucket[i].array[0] != expectedInitialCapacity {
			// Check if this slot corresponds to any of the keys
			found := false
			for _, key := range keys {
				if bfc.hash(key)%uint64(bfc.size) == uint64(i) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Untouched slot %d should have initial capacity %d, got %d", i, capacity, bfc.bucket[i].array[0])
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
	expectedCount := uint32(capacity) + uint32(numGoroutines*incrementCount)

	if bfc.bucket[index].array[0] != expectedCount {
		t.Errorf("Expected array[%d] to be %d, got %d", index, expectedCount, bfc.bucket[index].array[0])
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
			decrementWithRetry := bfc.DecrementWithRetryDecorator(bfc.Decrement)
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
			decrementWithRetry := bfc.DecrementWithRetryDecorator(bfc.Decrement)
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
		expectedCount := uint32(capacity) + uint32(numGoroutines*incrementCount)

		if bfc.bucket[index].array[0] != expectedCount {
			t.Errorf("For key '%s': Expected array[%d] to be %d, got %d", key, index, expectedCount, bfc.bucket[index].array[0])
		}
	}

	// Verify that untouched slots have initial capacity
	expectedInitialCapacity := uint32(capacity)
	for i := 0; i < int(size); i++ {
		if bfc.bucket[i].array[0] != expectedInitialCapacity {
			// Check if this slot corresponds to any of the keys
			found := false
			for _, key := range keys {
				if bfc.hash(key)%uint64(bfc.size) == uint64(i) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Untouched slot %d should have initial capacity %d, got %d", i, capacity, bfc.bucket[i].array[0])
			}
		}
	}
}
