package btree

import (
	"fmt"
	"io"
	"strings"
)

type Btree[KeyT ~int, ValueT any] struct {
	// The maximum number of child nodes of a node.
	order int
	root  *node[KeyT, ValueT]
}

// node per Knuth (wiki, m is order):
//
// 1. Every node has at most m children.
//
// 2. Every internal node has at least ⌈m/2⌉ children.
//
// 3. The root node has at least two children unless it is a leaf.
//
// 4. All leaves appear on the same level.
//
// 5 A non-leaf node with k children contains k−1 keys.
//
// The internal nodes have (at most) m-1 keys and m child nodes. The keys separate the child B-trees w.r.t. the range
// of the values in the sub-tree.
// The leaf nodes are the nodes that hold only "items". The item is the actual item that was inserted
// in the Btree itself.
// The leaf nodes have children (the items), but do not have keys.
type node[KeyT ~int, ValueT any] struct {
	// children nodes, either internal nodes, or the actually inserted items (keyValue structures).
	children []any
	// childrenCount counts how many children are set, so one does not need to sweep though all the children array.
	childrenCount int
	// keys separte children. There are len(children) - 1 keys at any moments.
	// Leaf nodes (the nodes that contain the the actually inserted items) do not have keys, the keys
	// array is set to nil.
	keys []KeyT
}

// keyValue is the struct that is
type keyValue[KeyT ~int, ValueT any] struct {
	key   KeyT
	value *ValueT
}

func New[KeyT ~int, ValueT any](order int) *Btree[KeyT, ValueT] {
	root := node[KeyT, ValueT]{
		children: make([]any, order),
		keys:     nil,
	}
	return &Btree[KeyT, ValueT]{
		order: order,
		root:  &root,
	}
}

func (b *Btree[KeyT, ValueT]) Find(key KeyT) *ValueT {
	if _, keyValue := b.root.findLeafNodeByKey(key); keyValue == nil {
		return nil
	} else {
		return keyValue.value
	}
}

// findLeafNodeByKey returns the leaf node that holds the value with seeked key, or the one that should
// hold such a value if it doesn't.
func (n *node[KeyT, ValueT]) findLeafNodeByKey(seekedKey KeyT) (*node[KeyT, ValueT], *keyValue[KeyT, ValueT]) {
	if n.isLeaf() {
		for _, c := range n.children {
			if c == nil {
				break
			}
			kv := c.(*keyValue[KeyT, ValueT])
			if kv.key == seekedKey {
				// Found.
				return n, kv
			}
		}
		// Not found.
		return n, nil
	} else {
		// There must always be at most m (order) children and len(children) - 1 keys that indicate which child
		// subtree has the keys in specific range. An example:
		//     0:10      1:20      2:30       -- keys
		// 0:n1      1:n2      2:n3     3:n4  -- child nodes
		// (-inf,10) |         |        |     -- range for node
		//           [10,20)   |        |
		//                     [20, 30) |
		//                              [30, +inf)
		for i, separatorKey := range n.keys {
			if separatorKey > seekedKey {
				// Too far, use the previous range (node).
				child := n.children[i].(*node[KeyT, ValueT])
				return child.findLeafNodeByKey(seekedKey)
			}
		}
		// Reached the last range.
		child := n.children[len(n.children)-1].(*node[KeyT, ValueT])
		return child.findLeafNodeByKey(seekedKey)
	}
}

func (b *Btree[KeyT, ValueT]) Insert(key KeyT, value ValueT) {
	leaf, kv := b.root.findLeafNodeByKey(key)
	if kv != nil {
		kv.value = &value
		return
	}
	assert(leaf.isLeaf(), "expected leaf for key %d", key)
	// https://en.wikipedia.org/wiki/B-tree#Insertion
	if b.isNodeFull(leaf) {
		panic("dupa")
	} else {
		leaf.insertChildInOrder(key, value)
	}
}

func (b *Btree[KeyT, ValueT]) isNodeFull(n *node[KeyT, ValueT]) bool {
	assert(n.childrenCount <= len(n.children), "children count exceeded children array length")
	assert(n.childrenCount <= b.order, "children count should not be larger than order")
	return n.childrenCount == b.order
}

// insertChildInOrder assumes there is space and there is no inserted item with `key`.
func (n *node[KeyT, ValueT]) insertChildInOrder(key KeyT, value ValueT) {
	assert(n.isLeaf(), "assumed leaf node")
	assert(n.childrenCount < len(n.children), "there is no spare capacity left in the array, insertChildInOrder should not be run at all.")
	insertAtIndex := 0
	// Find the index at which to insert the new key-value.
	for _, child := range n.children {
		if child == nil {
			break
		}
		childKv := child.(*keyValue[KeyT, ValueT])
		assert(childKv.key != key, "the case that the equal key is in the tree should have been already handled: childKv.key=%d , key=%d", childKv.key, key)
		if childKv.key > key {
			break
		}
		insertAtIndex++
	}
	// Here is time to insert KV. Move all the values to the right. The capacity should be already there.
	var curr, next any

	curr = n.children[insertAtIndex]
	n.children[insertAtIndex] = &keyValue[KeyT, ValueT]{key: key, value: &value}
	n.childrenCount++
	for i := insertAtIndex + 1; i < len(n.children); i++ {
		next = n.children[i]
		n.children[i] = curr
		curr = next
	}
}

// isLeaf says if the node is a leaf node, that is, a node that's children are the pointers to the actually
// stored values.
func (n *node[KeyT, ValueT]) isLeaf() bool {
	return n.keys == nil
}

func (b *Btree[KeyT, ValueT]) ValidityCheck() error {
	check := func(n *node[KeyT, ValueT]) error {
		if n.isLeaf() {
			var prevKey *KeyT
			for _, child := range n.children {
				kv := child.(*keyValue[KeyT, ValueT])
				if prevKey != nil {
					if !(*prevKey < kv.key) {
						return fmt.Errorf("for a child, prev key=%d, next key=%d", *prevKey, kv.key)
					}
				}
				prevKey = &kv.key
			}
		}
		return nil
	}
	return b.root.runRecursiveUntilError(check)
}

func (n *node[KeyT, ValueT]) runRecursiveUntilError(fun func(n *node[KeyT, ValueT]) error) error {
	if err := fun(n); err != nil {
		return err
	}
	for _, child := range n.children {
		if child != nil {
			if n2, ok := child.(*node[KeyT, ValueT]); ok {
				if err := n2.runRecursiveUntilError(fun); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (b *Btree[KeyT, ValueT]) Print(w io.Writer) {
	b.root.print(w, 0)
}

func (n *node[KeyT, ValueT]) print(w io.Writer, indent int) {
	spaces := strings.Repeat(" ", indent)
	fmt.Fprintf(w, "%schildren: %d isLeaf: %t\n", spaces, len(n.children), n.isLeaf())
	if n.isLeaf() {
		for i, child := range n.children {
			if child == nil {
				fmt.Fprintf(w, "%s[%d:nil] nil\n", spaces, i)
			} else {
				kv := child.(*keyValue[KeyT, ValueT])
				fmt.Fprintf(w, "%s[%d:%d] %s\n", spaces, i, kv.key, fmt.Sprint(kv.value))
			}
		}
	} else {
		for i, key := range n.keys {
			fmt.Fprintf(w, "%s%d", spaces, key)
			if i < (len(n.keys) - 1) {
				fmt.Fprintf(w, " | ")
			}
			fmt.Fprintf(w, "\n")
			for _, child := range n.children {
				n2 := child.(*node[KeyT, ValueT])
				n2.print(w, indent+1)
			}
		}
		for _, child := range n.children {
			kv := child.(*keyValue[KeyT, ValueT])
			fmt.Fprintf(w, "%s[k:%d] %s\n", spaces, kv.key, fmt.Sprint(kv.value))
		}
	}
}
