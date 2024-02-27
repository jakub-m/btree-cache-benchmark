package btree_test

import (
	"btree-cache-benchmark/btree"
	"fmt"
	"math/rand"
	"testing"
)

const (
	nValues = 100_000
)

var orders []int
var seed *rand.Rand

func init() {
	seed = rand.New(rand.NewSource(0))
	orders = []int{2, 3, 6, 10}
}

func BenchmarkInsertSequence(t *testing.B) {
	sequence := getSequence(nValues)
	for _, order := range orders {
		benchmarkInsert(t, sequence, order)
	}
}

func BenchmarkInsertSequenceShuffled(t *testing.B) {
	sequence := getSequence(nValues)
	shuffle(sequence)
	for _, order := range orders {
		benchmarkInsert(t, sequence, order)
	}
}

func BenchmarkInsertRandom(t *testing.B) {
	sequence := getRandomArray(nValues)
	for _, order := range orders {
		benchmarkInsert(t, sequence, order)
	}
}

func benchmarkInsert(t *testing.B, sequence []int, order int) {
	name := fmt.Sprintf("n_%d_order_%d", len(sequence), order)
	t.Run(name, func(b *testing.B) {
		t := btree.New[int, int](order)
		for _, value := range sequence {
			t.Insert(value, value)
		}
	})
}

func getSequence(n int) []int {
	s := make([]int, n)
	for i := range n {
		s[i] = i
	}
	return s
}

func getRandomArray(n int) []int {
	s := make([]int, n)
	for i := range n {
		s[i] = seed.Int()
	}
	return s
}

func shuffle[T any](s []T) {
	seed.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
}
