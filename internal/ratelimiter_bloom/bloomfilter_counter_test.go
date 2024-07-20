package ratelimiter_bloom

import (
	"sync"
	"testing"
)

func TestConcurrentIncrement(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"
	var wg sync.WaitGroup
	numGoroutines := 100
	incrementCount := 10

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
	capacity := uint32(20)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"

	bfc.Increment(key, 100) // Initialize with a known count

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bfc.Decrement(key)
		}()
	}

	wg.Wait()

	// Compute the index for the given key
	// index := bfc.hash(key) % uint64(bfc.size)
	// The expected count should be the initial capacity plus the initial increments minus the total decrements
	expectedCount := int(capacity) + 100 - numGoroutines

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
			bfc.Decrement(key) // Potentially making the key empty
		}()
	}

	wg.Wait()

	if !bfc.IsEmpty(key) {
		t.Errorf("Expected IsEmpty to return true")
	}
}

func TestConcurrentCheckCount(t *testing.T) {
	size := uint32(100)
	capacity := uint32(10)
	bfc := NewBloomFilterCounter(size, capacity)
	key := "test_key"

	bfc.Increment(key, 50) // Initialize with a known count

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementCount := 1

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bfc.Increment(key, incrementCount)
		}()
	}

	wg.Wait()

	// The expected count should be the initial count plus the total increments
	expectedCount := int(capacity) + 50 + numGoroutines*incrementCount

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

	// Increment each key concurrently
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
