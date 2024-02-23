package btree

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
)

type Btree[K cmp.Ordered, V any] struct {
	// The maximum number of child nodes of a node.
	order int
	// either innerNode or leafNode
	root node[K, V]
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
type node[K cmp.Ordered, V any] interface {
	// findLeafNodeByKey returns the leaf node that holds the value with seeked key, or the one that should
	// hold such a value if it doesn't.
	findLeafNodeByKey(key K) *leafNode[K, V]
	isRoot() bool
	runRecursiveUntilError(level int, fun func(level int, n node[K, V]) error) error
	print(w io.Writer, indent int)
}

////////////////////////////////////////
// Btree functions and methods
////////////////////////////////////////

func New[K ~int, V any](order int) *Btree[K, V] {
	root := newLeafNode[K, V]()
	return &Btree[K, V]{
		order: order,
		root:  root,
	}
}

func (b *Btree[K, V]) Find(key K) (V, bool) {
	if n := b.root.findLeafNodeByKey(key); n != nil {
		value, ok := n.values[key]
		return value, ok
	}
	var zero V
	return zero, false
}

func (b *Btree[K, V]) Insert(key K, value V) {
	leafNode := b.root.findLeafNodeByKey(key)
	assert(leafNode != nil, "there always must be some leaf node, not found for key %s", key)

	// https://en.wikipedia.org/wiki/B-tree#Insertion
	if !leafNode.isFull(b.order) {
		leafNode.insertAssumingHasSpace(key, value)
	}
	// The leaf node is full, so need to split.
	median := leafNode.medianKeyForChildrenAndKey(key)
	leftLeaf, rightLeaf := leafNode.splitAroundMedian(median)
	newInnerNode := newInnerNodeForSplitLeafs(leftLeaf, rightLeaf, median)
	if leafNode.isRoot() {
		b.replaceRoot(newInnerNode)
	} else {
		panic("dupa")
	}
}

func (b *Btree[K, V]) replaceRoot(n *innerNode[K, V]) {
	b.root = n
}

func (b *Btree[K, V]) IntegrityCheck() error {
	checkLeafSize := func(order int, n node[K, V]) error {
		leaf, ok := n.(*leafNode[K, V])
		if !ok {
			return nil
		}
		if len(leaf.values) > b.order {
			return fmt.Errorf("size of the leaf node is larger than the order")
		}
		return nil
	}
	return b.root.runRecursiveUntilError(0, checkLeafSize)
}

func (b *Btree[K, V]) Print(w io.Writer) {
	b.root.print(w, 0)
}

////////////////////////////////////////
// Inner node functions and methods
////////////////////////////////////////

// innerNode has children nodes that are either innerNodes or leafNodes.
type innerNode[K cmp.Ordered, V any] struct {
	children []node[K, V]
	// keys separate children. For m children there is always m-1 keys.
	keys   []K
	parent *innerNode[K, V]
}

func newInnerNodeForSplitLeafs[K cmp.Ordered, V any](leftLeaf, rightLeaf *leafNode[K, V], median K) *innerNode[K, V] {
	parent := &innerNode[K, V]{
		children: []node[K, V]{leftLeaf, rightLeaf},
		keys:     []K{median},
	}
	assert(leftLeaf.parent == nil, "after split, leaf parent should be null")
	leftLeaf.parent = parent
	assert(rightLeaf.parent == nil, "after split, leaf parent shoud be null")
	rightLeaf.parent = parent
	return parent
}

func (n *innerNode[K, V]) findLeafNodeByKey(seekedKey K) *leafNode[K, V] {
	// There must always be at most m (order) children and len(children) - 1 keys that indicate which child
	// subtree has the keys in specific range. An example:
	//     0:10      1:20      2:30       -- keys (separators), where in 2:30, the "2" is an index in the array, and "30" is the value.
	// 0:n1      1:n2      2:n3     3:n4  -- child nodes, where in 2:n3, the "2" is an index, and "n3" is an identifier of the node.
	// (-inf,10) |         |        |     -- range for node
	//           [10,20)   |        |
	//                     [20, 30) |
	//                              [30, +inf)
	foundNodeIndex := len(n.keys) // if no key found, use the last range
	for i, separator := range n.keys {
		if separator > seekedKey {
			foundNodeIndex = i
		}
	}
	// Reached the last range.
	assert(foundNodeIndex < len(n.children), "found node index is outside children range")
	return n.children[foundNodeIndex].findLeafNodeByKey(seekedKey)
}

func (n *innerNode[K, V]) isRoot() bool {
	return n.parent == nil
}

func (n *innerNode[K, V]) runRecursiveUntilError(level int, fun func(level int, n node[K, V]) error) error {
	if err := fun(level, n); err != nil {
		return err
	}
	for _, child := range n.children {
		if err := child.runRecursiveUntilError(level+1, fun); err != nil {
			return err
		}
	}
	return nil
}

func (n *innerNode[K, V]) print(w io.Writer, indent int) {
	spaces := strings.Repeat(" ", indent)
	for i, key := range n.keys {
		fmt.Fprintf(w, "%s%v", spaces, key)
		if i < (len(n.keys) - 1) {
			fmt.Fprintf(w, " | ")
		}
	}
	fmt.Fprintf(w, "\n")
	for _, child := range n.children {
		child.print(w, indent+1)
	}
}

////////////////////////////////////////
// Leaf node functions and methods
////////////////////////////////////////

// leafNode contains no children, but arbitrary values stored under keys.
type leafNode[K cmp.Ordered, V any] struct {
	values map[K]V
	parent *innerNode[K, V]
}

func newLeafNode[K cmp.Ordered, V any]() *leafNode[K, V] {
	return &leafNode[K, V]{
		values: make(map[K]V),
	}
}

func (n *leafNode[K, V]) findLeafNodeByKey(seekedKey K) *leafNode[K, V] {
	return n
}

func (n *leafNode[K, V]) isRoot() bool {
	return n.parent == nil
}

func (n *leafNode[K, V]) insertAssumingHasSpace(key K, value V) {
	n.values[key] = value
}

func (n *leafNode[K, V]) isFull(order int) bool {
	assert(len(n.values) < order, "length of values array must be smaller than order")
	return len(n.values) == (order - 1)
}

// medianKeyForChildrenAndKey return the median out of children elements and the new key.
func (n *leafNode[K, V]) medianKeyForChildrenAndKey(key K) K {
	keys := make([]K, 0, len(n.values)+1)
	for k := range n.values {
		keys = append(keys, k)
	}
	keys[len(n.values)] = key // insert last key
	slices.Sort(keys)
	return keys[len(keys)/2]
}

func (n *leafNode[K, V]) splitAroundMedian(median K) (*leafNode[K, V], *leafNode[K, V]) {
	left, right := newLeafNode[K, V](), newLeafNode[K, V]()
	for k, v := range n.values {
		if k < median {
			left.values[k] = v
		} else {
			right.values[k] = v
		}
	}
	return left, right
}

func (n *leafNode[K, V]) runRecursiveUntilError(level int, fun func(level int, n node[K, V]) error) error {
	if err := fun(level, n); err != nil {
		return err
	}
	return nil
}

func (n *leafNode[K, V]) print(w io.Writer, indent int) {
	spaces := strings.Repeat(" ", indent)
	for k, v := range n.values {
		fmt.Fprintf(w, "%s[%v]:%v\n", spaces, k, v)
	}
}

/// func (n *node[K, V]) createLeftNodeAfterSplit(median K) *node[K, V] {
/// 	leftChildren := make([]any, len(n.children))
/// 	rightChildren := make([]any, len(n.children))
/// 	if n.isLeaf() {
/// 		for _, c := range n.children {
/// 			kv := c.(*keyValue[K, V])
/// 			assert(kv.key != median, "median should not be equal to any key value")
/// 			if kv.key < median {
/// 				leftChildren = append(leftChildren, c)
/// 			} else {
/// 				rightChildren = append(rightChildren, c)
/// 			}
/// 		}
/// 	} else {
///
/// 	}
///
/// 	// return &node[K, V]{
/// 	// 	children: ...,
/// 	// 	childrenCount, ...
/// 	// 	keys: ...
/// 	// }
/// }

///////////////////////////////////////////////////////

/// // insertChildInOrder assumes there is space and there is no inserted item with `key`.
/// func (n *node[K, V]) insertChildInOrder(key K, value V) {
/// 	assert(n.isLeaf(), "assumed leaf node")
/// 	assert(n.childrenCount < len(n.children), "there is no spare capacity left in the array, insertChildInOrder should not be run at all.")
/// 	insertAtIndex := 0
/// 	// Find the index at which to insert the new key-value.
/// 	for _, child := range n.children {
/// 		if child == nil {
/// 			break
/// 		}
/// 		childKv := child.(*keyValue[K, V])
/// 		assert(childKv.key != key, "the case that the equal key is in the tree should have been already handled: childKv.key=%d , key=%d", childKv.key, key)
/// 		if childKv.key > key {
/// 			break
/// 		}
/// 		insertAtIndex++
/// 	}
/// 	// Here is time to insert KV. Move all the values to the right. The capacity should be already there.
/// 	var curr, next any
///
/// 	curr = n.children[insertAtIndex]
/// 	n.children[insertAtIndex] = &keyValue[K, V]{key: key, value: &value}
/// 	n.childrenCount++
/// 	for i := insertAtIndex + 1; i < len(n.children); i++ {
/// 		next = n.children[i]
/// 		n.children[i] = curr
/// 		curr = next
/// 	}
/// }
