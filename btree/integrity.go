package btree

import (
	"cmp"
	"fmt"
	"slices"
)

func (b *Btree[K, V]) IntegrityCheck() error {
	keyPerNodeChecker := newKeyPerNodeChecker[K, V](b.root)
	chained := chainIntegrityCheck[K, V](
		b.integrityCheckLeafSize,
		b.integrityCheckKeyAndChildrenLen,
		b.integrityCheckAllButRootHaveParent,
		b.integrityCheckParentPointsCorrectly,
		keyPerNodeChecker.check,
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

func (b *Btree[K, V]) integrityCheckLeafSize(level int, n node[K, V]) error {
	leaf, ok := n.(*leafNode[K, V])
	if !ok {
		return nil
	}
	if len(leaf.values) > b.order {
		return fmt.Errorf("size of the leaf node is larger than the order")
	}
	return nil
}

func (b *Btree[K, V]) integrityCheckKeyAndChildrenLen(level int, n node[K, V]) error {
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

func (b *Btree[K, V]) integrityCheckAllButRootHaveParent(level int, n node[K, V]) error {
	if level == 0 {
		if n.getParent() != nil {
			return fmt.Errorf("expected root to have no parent")
		}
	} else {
		if n.getParent() == nil {
			return fmt.Errorf("expected non-root to have parent")
		}
	}
	return nil
}

func (b *Btree[K, V]) integrityCheckParentPointsCorrectly(level int, n node[K, V]) error {
	switch t := n.(type) {
	case *innerNode[K, V]:
		{
			for _, c := range t.children {
				if c.getParent() != n {
					return fmt.Errorf("parent of child node does not point to correct parent")
				}
			}
		}
	}
	return nil
}

type keyPerNodeChecker[K cmp.Ordered, V any] struct {
	keysPerNode map[node[K, V]][]K
}

func newKeyPerNodeChecker[K cmp.Ordered, V any](n node[K, V]) *keyPerNodeChecker[K, V] {
	c := &keyPerNodeChecker[K, V]{
		keysPerNode: make(map[node[K, V]][]K),
	}
	c.collectKeysPerNode(n)
	return c
}

func (c *keyPerNodeChecker[K, V]) collectKeysPerNode(n node[K, V]) {
	switch t := n.(type) {
	case *leafNode[K, V]:
		keys := []K{}
		for k := range t.values {
			keys = append(keys, k)
		}
		assert(c.keysPerNode[n] == nil)
		c.keysPerNode[n] = keys
	case *innerNode[K, V]:
		keys := []K{}
		for _, child := range t.children {
			c.collectKeysPerNode(child)
			subKeys := c.keysPerNode[child]
			assert(subKeys != nil)
			keys = append(keys, subKeys...)
		}
		assert(c.keysPerNode[n] == nil)
		c.keysPerNode[n] = keys
	}
}

func (c *keyPerNodeChecker[K, V]) check(level int, n node[K, V]) error {
	inner, ok := n.(*innerNode[K, V])
	if !ok {
		return nil
	}
	for i, child := range inner.children {
		keysForChild := c.keysPerNode[child]
		assert(keysForChild != nil)
		leftmost := i == 0
		rightmost := i == len(inner.keys)
		minKey := slices.Min(keysForChild)
		maxKey := slices.Max(keysForChild)
		if !leftmost && !(minKey >= inner.keys[i-1]) {
			return fmt.Errorf("bad min key")
		}
		if !rightmost && !(maxKey < inner.keys[i]) {
			return fmt.Errorf("mad max key")
		}
	}
	return nil
}
