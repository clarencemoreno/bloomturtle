package main

import (
	"fmt"
	"sync"
)

type ArrayWrapper struct {
	data *[5]int
	mu   sync.RWMutex
}

func (aw *ArrayWrapper) Replace(newArray *[5]int) {
	fmt.Println("Locking for Replace")
	aw.mu.Lock()
	defer func() {
		fmt.Println("Unlocking after Replace")
		aw.mu.Unlock()
	}()
	aw.data = newArray
}

func (aw *ArrayWrapper) Get() *[5]int {
	fmt.Println("Locking for Get")
	aw.mu.RLock()
	defer func() {
		fmt.Println("Unlocking after Get")
		aw.mu.RUnlock()
	}()
	return aw.data
}
