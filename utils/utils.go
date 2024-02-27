package utils

import "math/rand"

var seed *rand.Rand

func init() {
	seed = rand.New(rand.NewSource(0))
}

func GetSequenceRange(n int) []int {
	s := make([]int, n)
	for i := range n {
		s[i] = i
	}
	return s
}

func GetRandomArray(n int) []int {
	s := make([]int, n)
	for i := range n {
		s[i] = seed.Int()
	}
	return s
}

func Shuffle[T any](s []T) {
	seed.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
}
