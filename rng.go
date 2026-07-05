package main

import "math/rand"

func randIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}
