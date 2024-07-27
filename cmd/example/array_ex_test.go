package main

import (
	"sync"
	"testing"
)

func TestArrayWrapperConcurrentReplace(t *testing.T) {
	aw := &ArrayWrapper{
		data: &[5]int{1, 2, 3, 4, 5},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Read the current array
			oldArray := aw.Get()

			// Clone the array to a slice
			slice := make([]int, len(oldArray))
			copy(slice, oldArray[:])

			// Modify the slice: either add or remove an item
			if index%2 == 0 && len(slice) > 1 {
				// Remove an item (remove the first item in this case)
				slice = slice[1:]
			} else {
				// Add a new item to the end of the slice
				slice = append(slice, index+10)
			}

			// Convert the slice back to an array
			var newArray [5]int
			copy(newArray[:], slice)

			// Replace the old array with the modified new array
			aw.Replace(&newArray)
		}(i)
	}

	wg.Wait()

	// Check the final state of the array
	finalArray := aw.Get()

	// Determine what the expected final array should be
	// If you add items, they will be the last ones added; if you remove items,
	// ensure the final array reflects the results of those operations
	expectedArray := &[5]int{0, 0, 0, 0, 0} // This should be adjusted based on your expected final result

	if *finalArray != *expectedArray {
		t.Errorf("Expected final array %v, but got %v", *expectedArray, *finalArray)
	} else {
		t.Logf("Final array is correctly %v", *finalArray)
	}
}
