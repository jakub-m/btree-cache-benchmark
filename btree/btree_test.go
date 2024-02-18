package btree_test

import (
	"btree-cache-benchmark/btree"
	"fmt"
	"testing"
)

func TestLotsOfSequentialInsertions(t *testing.T) {
	for _, order := range []int{2, 3, 4, 10} {
		t.Run(fmt.Sprintf("order %d", order), func(t *testing.T) {
			b := btree.New(order)
			for value := range 1000 {
				b.Insert(value, value)
			}
		})
	}

}
