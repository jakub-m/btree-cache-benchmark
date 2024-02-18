package btree_test

import (
	"btree-cache-benchmark/btree"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLotsOfSequentialInsertions(t *testing.T) {
	n := 1000
	for _, order := range []int{2, 3, 4, 10} {
		t.Run(fmt.Sprintf("order %d", order), func(t *testing.T) {
			t.Parallel()
			b := btree.New[int, int](order)
			for i := range n {
				b.Insert(i, i)
			}
			assert.NoError(t, b.ValidityCheck())
			assert.Nil(t, b.Find(-1))
			for i := range n {
				found := b.Find(i)
				assert.NotNilf(t, found, "expected to find key %d, got nil", i)
				assert.Equal(t, *found, i, "expected concrete value, got other")
			}
			assert.Nil(t, b.Find(-1), "expected -1 to not be found")
			assert.Nilf(t, b.Find(n), "expected %d to not be found", n)
		})
	}
}
