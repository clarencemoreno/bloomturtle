package ratelimiter_bloom

import (
	"hash/fnv"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// BloomFilterCounter implements a Bloom filter counter.
type BloomFilterCounter struct {
	array    []uint32
	size     uint32
	capacity uint32
	mu       sync.Mutex
}

// NewBloomFilterCounter creates a new Bloom filter counter with the specified size.
// The default capacity is set to 10 if not specified.
func NewBloomFilterCounter(size uint32, capacity ...uint32) *BloomFilterCounter {
	cap := uint32(10)
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	bfc := &BloomFilterCounter{
		array:    make([]uint32, size),
		size:     size,
		capacity: cap,
	}
	for i := range bfc.array {
		bfc.array[i] = cap
	}
	return bfc
}

// Function type for mutator functions
type MutatorFunc func(key string, count int)

// IncrementWithRetry increments the counter for the given key by the specified count with retry logic.
func (bfc *BloomFilterCounter) IncrementWithRetry(key string, count int) {
	index := bfc.hash(key) % uint64(bfc.size)
	var attempt int
	for attempt < 3 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Millisecond) // Exponential backoff can be applied here
		}
		if atomic.CompareAndSwapUint32(&bfc.array[index], bfc.array[index], bfc.array[index]+uint32(count)) {
			return
		}
		attempt++
	}
	// Fallback to locking mechanism after max attempts
	bfc.mu.Lock()
	defer bfc.mu.Unlock()
	bfc.array[index] += uint32(count)
}

// DecrementWithRetry attempts to decrement the counter for the given key with retry logic.
// It prevents the counter from going below zero.
func (bfc *BloomFilterCounter) DecrementWithRetry(key string, count int) {
	index := bfc.hash(key) % uint64(bfc.size)
	var attempt int
	for attempt < 3 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Millisecond) // Exponential backoff can be applied here
		}
		// Read the current value
		currentValue := atomic.LoadUint32(&bfc.array[index])

		// If the value is already zero, just return
		if currentValue == 0 {
			return
		}

		// Calculate the new value after decrementing
		newValue := currentValue - uint32(count)
		if newValue > currentValue { // Ensure we do not have underflow
			newValue = 0
		}

		// Attempt to decrement the value
		if atomic.CompareAndSwapUint32(&bfc.array[index], currentValue, newValue) {
			return
		}
		attempt++
	}

	// Fallback to locking mechanism after max attempts
	bfc.mu.Lock()
	defer bfc.mu.Unlock()

	// Recheck value after acquiring the lock
	if bfc.array[index] == 0 {
		return
	}

	// Calculate the new value after decrementing
	newValue := bfc.array[index] - uint32(count)
	if newValue > bfc.array[index] { // Ensure we do not have underflow
		newValue = 0
	}
	bfc.array[index] = newValue
}

// IsEmpty checks if the counter for the given key is greater than 0.
func (bfc *BloomFilterCounter) IsEmpty(key string) bool {
	index := bfc.hash(key) % uint64(bfc.size)
	return bfc.array[index] == 0
}

// CheckCount checks the count for the given key.
func (bfc *BloomFilterCounter) CheckCount(key string) int {
	index := bfc.hash(key) % uint64(bfc.size)
	return int(bfc.array[index])
}

// hash uses a more robust hash function (FNV-1a)
func (bfc *BloomFilterCounter) hash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

// Decorator function that adds retry logic
func (bfc *BloomFilterCounter) IncrementWithRetryDecorator(fn MutatorFunc) MutatorFunc {
	return func(key string, count int) {
		var attempt int
		maxAttempts := 3
		for attempt < maxAttempts {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * time.Millisecond) // Exponential backoff can be applied here
			}
			if atomic.CompareAndSwapUint32(&bfc.array[bfc.hash(key)%uint64(bfc.size)], bfc.array[bfc.hash(key)%uint64(bfc.size)], bfc.array[bfc.hash(key)%uint64(bfc.size)]+uint32(count)) {
				return
			}
			attempt++
		}
		// Fallback to original function after max attempts
		fn(key, count)
	}
}

func (bfc *BloomFilterCounter) DecrementWithRetryDecorator(fn MutatorFunc) MutatorFunc {
	return func(key string, count int) {

		index := bfc.hash(key) % uint64(bfc.size)
		var attempt int
		maxAttempts := 3

		for attempt < maxAttempts {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * time.Millisecond) // Exponential backoff can be applied here
			}
			// Read the current value
			currentValue := atomic.LoadUint32(&bfc.array[index])

			// If the value is already zero, just return
			if currentValue == 0 {
				return
			}

			// Calculate the new value after decrementing
			newValue := currentValue - uint32(count)
			if newValue > currentValue { // Ensure we do not have underflow
				newValue = 0
			}

			// Attempt to decrement the value
			if atomic.CompareAndSwapUint32(&bfc.array[index], currentValue, newValue) {
				return
			}
			attempt++
		}

		// Fallback to original function after max attempts
		fn(key, count)
	}
}

// Decrement decrements the counter for the given key by the specified count.
func (bfc *BloomFilterCounter) Decrement(key string, count int) {
	index := bfc.hash(key) % uint64(bfc.size)
	bfc.mu.Lock()
	defer bfc.mu.Unlock()
	// No value of uint32 is less than zero
	if bfc.array[index] > 0 {
		bfc.array[index] -= uint32(count)
	}
	// Zero or less: no op
}

// Increment increments the counter for the given key by the specified count.
func (bfc *BloomFilterCounter) Increment(key string, count int) {
	index := bfc.hash(key) % uint64(bfc.size)
	bfc.mu.Lock()
	defer bfc.mu.Unlock()
	if bfc.array[index] < math.MaxUint32 {
		bfc.array[index] += uint32(count)
	}
}
