package btree

import (
	"cmp"
	"fmt"
	"slices"
)

// TODO add integrity check checking values in sub-trees
func (b *Btree[K, V]) IntegrityCheck() error {
	chained := chainIntegrityCheck[K, V](
		b.integrityCheckLeafSize,
		b.integrityCheckKeyAndChildrenLen,
	)
	return b.root.runRecursiveUntilError(0, chained)
}

func chainIntegrityCheck[K cmp.Ordered, V any](funcs ...func(level int, n node[K, V]) error) func(level int, n node[K, V]) error {
	return func(level int, n node[K, V]) error {
		for _, f := range funcs {
			if err := f(level, n); err != nil {
				return err
			}
		}
		return nil
	}
}

func (b *Btree[K, V]) integrityCheckLeafSize(order int, n node[K, V]) error {
	leaf, ok := n.(*leafNode[K, V])
	if !ok {
		return nil
	}
	if len(leaf.values) > b.order {
		return fmt.Errorf("size of the leaf node is larger than the order")
	}
	return nil
}

func (b *Btree[K, V]) integrityCheckKeyAndChildrenLen(order int, n node[K, V]) error {
	inner, ok := n.(*innerNode[K, V])
	if !ok {
		return nil
	}
	if len(inner.children) != len(inner.keys)+1 {
		return fmt.Errorf("len children (%d) != len keys + 1 (%d)", len(inner.children), len(inner.keys))
	}
	if !slices.IsSorted(inner.keys) {
		return fmt.Errorf("keys are not sorted: %v", inner.keys)
	}
	return nil
}
