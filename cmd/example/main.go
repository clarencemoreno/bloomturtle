package main

import (
	"fmt"

	"github.com/clarencemoreno/bloomturtle"
)

func main() {
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint { /* hash function 1 */ return 0 },
		func(data []byte) uint { /* hash function 2 */ return 1 },
	}

	bf := bloomturtle.NewBloomFilter(1000, hashFuncs)

	bf.Add([]byte("hello"))
	fmt.Println(bf.Contains([]byte("hello"))) // Output: true
	fmt.Println(bf.Contains([]byte("world"))) // Output: false
}
