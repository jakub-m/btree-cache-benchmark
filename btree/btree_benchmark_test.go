package btree_test

import (
	"btree-cache-benchmark/btree"
	"btree-cache-benchmark/utils"
	"fmt"
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
var sequenceTypes []string

func init() {
	orders = []int{2, 3, 6, 10, 23}
	sequenceTypes = []string{
		sequenceTypeRange,
		sequenceTypeShuffledRange,
		// sequenceTypeRandom,
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
		return utils.GetSequenceRange(n)
	case sequenceTypeShuffledRange:
		s := utils.GetSequenceRange(n)
		utils.Shuffle(s)
		return s
	case sequenceTypeRandom:
		return utils.GetRandomArray(n)
	}
	panic("bad type: " + t)
}
