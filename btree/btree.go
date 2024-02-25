// Schematically the insertion looks as follows, for a B-tree of order 2.
//
//	   . 15    .   25  .
//	     |   20,21 | 30,40
//	---------------------------
//	50!                           // Insert new value 50
//	---------------------------
//	             30,40,50      !
//	---------------------------
//				  . 40 .          // 40 is the median used to split leafs.
//				30   | 40,50      // New left and new right and median of 40.
//	---------------------------
//	 . 15   .    25 . 40    .  !  // Try he parent is already full, needs to split again
//	   |  20,21  |  30 |  40,50
//	---------------------------
//	    .     25     .            // The split uses 15,25,40, so 25 is new median with two sub-trees
//	 .15 .     |   . 40  .        // the new node with 25 as a key is added to the parent.
//	   | 20,21 | 30  | 40,50
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
	getParent() *innerNode[K, V]
	setParent(parent *innerNode[K, V])
	runRecursiveUntilError(level int, fun func(level int, n node[K, V]) error) error
	// The returned node is (optional) new root node.
	// insertNodesToParentRec(child, left, right node[K, V], order int, median K) *innerNode[K, V]
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
	// https://en.wikipedia.org/wiki/B-tree#Insertion
	leafNode := b.root.findLeafNodeByKey(key)
	assert(leafNode != nil, "there always must be some leaf node, not found for key %s", key)
	leafNode.insertIgnoringOrder(key, value)
	if !leafNode.isOverflow(b.order) {
		return
	}
	left, right, median := leafNode.splitAroundMedian()
	if newRoot := b.replaceNodeWithTwoNodesAndSeparatorRec(leafNode, left, right, median); newRoot != nil {
		b.root = newRoot
	}
}

// replaceNodeWithTwoNodesAndSeparatorRec does not care about order. Optionally, returns new root node.
func (b *Btree[K, V]) replaceNodeWithTwoNodesAndSeparatorRec(childToRemove, left, right node[K, V], separator K) *innerNode[K, V] {
	parent := childToRemove.getParent()
	if parent == nil {
		newParent := &innerNode[K, V]{
			children: []node[K, V]{left, right},
			keys:     []K{separator},
		}
		left.setParent(newParent)
		right.setParent(newParent)
		return newParent
	}
	parent.expandAtChild(childToRemove, left, right, separator)
	if !parent.isOverflow(b.order) {
		return nil
	}
	newLeft, newRight, newMedian := parent.splitAroundMedian()
	assert(newLeft.getParent() == nil, "new split left should have nil parent")
	assert(newRight.getParent() == nil, "new split right should have nil parent")
	return b.replaceNodeWithTwoNodesAndSeparatorRec(parent, newLeft, newRight, newMedian)
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
	// Key i is the key after child i, like:
	//   child[0], key[0], child[1], key[1], child[2], key[2], child[3]
	keys   []K
	parent *innerNode[K, V]
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

func (n *innerNode[K, V]) isOverflow(order int) bool {
	assert(len(n.children) <= order+1, "there should be no path that results in child len > one more than order, len(children)=%d, order=%d", len(n.children), order)
	return len(n.children) > order
}

func (n *innerNode[K, V]) expandAtChild(childToRemove, left, right node[K, V], separator K) {
	i := slices.Index(n.children, childToRemove)
	if i == -1 {
		panic("BUG! Could not find child!")
	}
	// This can be optimized to not delete but replace in place with left node.
	n.children = slices.Delete(n.children, i, i+1)
	n.children = slices.Insert(n.children, i, left, right)
	n.keys = slices.Insert(n.keys, i, separator)
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
	fmt.Fprintf(w, "%s--\n", spaces)
	for i, key := range n.keys {
		n.children[i].print(w, indent+1)
		fmt.Fprintf(w, "%s%v:\n", spaces, key)
	}
	n.children[len(n.children)-1].print(w, indent+1)
	fmt.Fprintf(w, "%s--\n", spaces)
}

func (n *innerNode[K, V]) splitAroundMedian() (*innerNode[K, V], *innerNode[K, V], K) {
	assert(slices.IsSorted(n.keys), "expected keys to be sorted, was: %v", n.keys)
	iMedian := len(n.keys) / 2
	medianValue := n.keys[iMedian]
	leftChildren := slices.Clone(n.children[:iMedian+1]) // clone to allow GC collecting n.children
	leftKeys := slices.Clone(n.keys[:iMedian])
	rightChildren := slices.Clone(n.children[iMedian+1:])
	rightKeys := slices.Clone(n.keys[iMedian+1:])
	newLeft := &innerNode[K, V]{
		children: leftChildren,
		keys:     leftKeys,
	}
	newRight := &innerNode[K, V]{
		children: rightChildren,
		keys:     rightKeys,
	}
	return newLeft, newRight, medianValue
}

func (n *innerNode[K, V]) getParent() *innerNode[K, V] {
	return n.parent
}

func (n *innerNode[K, V]) setParent(p *innerNode[K, V]) {
	assert(n.parent == nil, "there should be no code path setting non-nil parent")
	n.parent = p
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

func (n *leafNode[K, V]) isOverflow(order int) bool {
	return len(n.values) > order
}

func (n *leafNode[K, V]) getParent() *innerNode[K, V] {
	return n.parent
}

func (n *leafNode[K, V]) setParent(p *innerNode[K, V]) {
	assert(n.parent == nil, "there should be no code path setting non-nil parent")
	n.parent = p
}

func (n *leafNode[K, V]) insertIgnoringOrder(key K, value V) {
	n.values[key] = value
}

func (n *leafNode[K, V]) splitAroundMedian() (*leafNode[K, V], *leafNode[K, V], K) {
	median := n.medianKey()
	left, right := newLeafNode[K, V](), newLeafNode[K, V]()
	insertToLeftOrRight := func(k K, v V) {
		if k < median {
			left.values[k] = v
		} else {
			right.values[k] = v
		}
	}
	for k, v := range n.values {
		insertToLeftOrRight(k, v)
	}
	return left, right, median
}

func (n *leafNode[K, V]) medianKey() K {
	keys := make([]K, 0, len(n.values))
	for k := range n.values {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys[len(keys)/2]
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
