package main

import (
	"hash/fnv"
	"testing"
)

func BenchmarkFNV32a(b *testing.B) {
	key := []byte("example key")
	for i := 0; i < b.N; i++ {
		h := fnv.New32a()
		h.Write(key)
		_ = h.Sum32()
	}
}

func BenchmarkFNV64a(b *testing.B) {
	key := []byte("example key")
	for i := 0; i < b.N; i++ {
		h := fnv.New64a()
		h.Write(key)
		_ = h.Sum64()
	}
}
