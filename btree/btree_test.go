package btree_test

import (
	"btree-cache-benchmark/btree"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOne(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(10, 42)
	res := b.Find(10)
	assert.NotNil(t, res)
	assert.Equal(t, 42, *res)
	b.Print(os.Stderr)
}

func TestInsertTwoOutOfOrder(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(20, 120)
	b.Insert(10, 110)
	assert.NoError(t, b.ValidityCheck())
	res := b.Find(10)
	assert.NotNil(t, res)
	assert.Equal(t, 110, *res)
	b.Print(os.Stderr)
}
func TestInsertInOfOrder(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(10, 110)
	b.Insert(20, 120)
	assert.NoError(t, b.ValidityCheck())
	res := b.Find(10)
	assert.NotNil(t, res)
	assert.Equal(t, 110, *res)
	b.Print(os.Stderr)
}

func TestLotsOfSequentialInsertions(t *testing.T) {
	n := 1000
	for _, order := range []int{2, 3, 4, 10} {
		order := order
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
