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

const (
	sequenceTypeRange         = "range"
	sequenceTypeShuffledRange = "shuffledRange"
	sequenceTypeRandom        = "random"
)

var orders []int
var seed *rand.Rand
var sequenceTypes []string

func init() {
	seed = rand.New(rand.NewSource(0))
	orders = []int{2, 3, 6, 10, 23}
	sequenceTypes = []string{
		sequenceTypeRange,
		sequenceTypeShuffledRange,
		sequenceTypeRandom,
	}
}

func BenchmarkInsert(t *testing.B) {
	for _, order := range orders {
		for _, s := range sequenceTypes {
			runBenchmarkForInsert(t, s, order)
		}
	}
}

func runBenchmarkForInsert(t *testing.B, sequenceType string, order int) {
	name := fmt.Sprintf("n:%d_order:%d_seq:%s", nValues, order, sequenceType)
	sequence := getSequence(nValues, sequenceType)
	t.Run(name, func(b *testing.B) {
		for range b.N {
			t := btree.New[int, int](order)
			for _, value := range sequence {
				t.Insert(value, value)
			}
		}
	})
}

func getSequence(n int, t string) []int {
	switch t {
	case sequenceTypeRange:
		return getSequenceRange(n)
	case sequenceTypeShuffledRange:
		s := getSequenceRange(n)
		shuffle(s)
		return s
	case sequenceTypeRandom:
		return getRandomArray(n)
	}
	panic("bad type: " + t)
}
func getSequenceRange(n int) []int {
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
