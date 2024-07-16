package main

import (
	"context"
	"testing"

	"github.com/clarencemoreno/bloomturtle"
)

func TestMainFunction(t *testing.T) {
	hashFuncs := []func([]byte) uint{
		func(data []byte) uint {
			if len(data) > 0 {
				return uint(data[0])
			}
			return 0
		},
		func(data []byte) uint {
			if len(data) > 1 {
				return uint(data[1])
			}
			return 0
		},
	}

	rl := bloomturtle.NewRateLimiter(1000, hashFuncs, 50)
	defer rl.Shutdown(context.Background())

	data := []byte("data0")
	err := rl.Add(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contains, err := rl.Contains(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains {
		t.Errorf("expected data0 to be contained in the rate limiter")
	}

	err = rl.Add([]byte("data100"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contains, err = rl.Contains([]byte("data100"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains {
		t.Errorf("expected data100 to be contained in the rate limiter")
	}
}
