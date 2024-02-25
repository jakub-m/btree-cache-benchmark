package btree_test

import (
	"btree-cache-benchmark/btree"
	"cmp"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOne(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(10, 42)
	assertFound(t, b, 10, 42)
	b.Print(os.Stderr)
}

func TestInsertTwoOutOfOrder(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(20, 120)
	b.Insert(10, 110)
	assert.NoError(t, b.IntegrityCheck())
	assertFound(t, b, 10, 110)
	assertFound(t, b, 20, 120)
	b.Print(os.Stderr)
}
func TestInsertInOrder(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(10, 110)
	b.Insert(20, 120)
	b.Print(os.Stderr)

	assert.NoError(t, b.IntegrityCheck())
	assertFound(t, b, 10, 110)
	assertFound(t, b, 20, 120)
}

func TestInsertOverOrder(t *testing.T) {
	b := btree.New[int, int](2)
	b.Insert(10, 110)
	b.Insert(20, 120)
	b.Insert(30, 130)
	b.Print(os.Stderr)
	b.IntegrityCheck()

	assertFound(t, b, 10, 110)
	assertFound(t, b, 20, 120)
	assertFound(t, b, 30, 130)
}

func TestInsertTwiceOverOrder(t *testing.T) {
	b := btree.New[int, int](2)
	for _, kv := range [][2]int{
		{10, 110},
		{20, 120},
		{30, 130},
		{40, 140},
	} {

		b.Insert(kv[0], kv[1])
		fmt.Fprintf(os.Stderr, "inserted %d\n", kv[0])
		b.Print(os.Stderr)
	}
	b.IntegrityCheck()

	assertFound(t, b, 10, 110)
	assertFound(t, b, 20, 120)
	assertFound(t, b, 30, 130)
	assertFound(t, b, 40, 140)
}

func TestLotsOfSequentialInsertions(t *testing.T) {
	n := 1000
	for _, order := range []int{2, 3, 4, 10} {
		order := order
		t.Run(fmt.Sprintf("order %d", order), func(t *testing.T) {
			// t.Parallel()
			b := btree.New[int, int](order)
			for i := range n {
				b.Insert(i, i)
			}
			assert.NoError(t, b.IntegrityCheck())
			assertNotFound(t, b, -1)
			for i := range n {
				assertFound(t, b, i, i)
			}
			assertNotFound(t, b, -1)
			assertNotFound(t, b, n)
		})
	}
}

func TestLotsOfRandomInsertions(t *testing.T) {
	t.Skip("here should have lots of random insertions")
}

func assertFound[K cmp.Ordered, V any](t *testing.T, b *btree.Btree[K, V], key K, expected V) {
	t.Helper()
	actual, ok := b.Find(key)
	assert.True(t, ok, "value not found for key %s", key)
	assert.Equal(t, expected, actual, "value differs for key %s", key)
}

func assertNotFound[K cmp.Ordered, V any](t *testing.T, b *btree.Btree[K, V], key K) {
	_, ok := b.Find(key)
	assert.False(t, ok, "value found for key %s", key)
}
