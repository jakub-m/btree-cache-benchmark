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
	"sort"
	"strings"
)

type Btree[K cmp.Ordered, V any] struct {
	// The maximum number of child nodes of a node.
	order int
	// either innerNode or leafNode
	root             node[K, V]
	accessCounter    accessCounter
	rebalanceCounter rebalanceCounter
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
	countAccess()
}

////////////////////////////////////////
// Btree functions and methods
////////////////////////////////////////

func New[K ~int, V any](order int) *Btree[K, V] {
	ac := dummyAccessCounter
	root := newLeafNode[K, V](ac)
	return &Btree[K, V]{
		order:         order,
		root:          root,
		accessCounter: ac,
	}
}

// SetAccessCounter must be called right after New.
func (b *Btree[K, V]) SetAccessCounter(ac accessCounter) {
	b.accessCounter = ac
	b.root.(*leafNode[K, V]).accessCounter = ac
}

func (b *Btree[K, V]) SetRebalanceCounter(rc rebalanceCounter) {
	b.rebalanceCounter = rc
}

func (b *Btree[K, V]) Find(key K) (V, bool) {
	if n := b.root.findLeafNodeByKey(key); n != nil {
		return n.getValue(key)
	}
	var zero V
	return zero, false
}

func (b *Btree[K, V]) Insert(key K, value V) {
	// https://en.wikipedia.org/wiki/B-tree#Insertion
	leafNode := b.root.findLeafNodeByKey(key)
	assert(leafNode != nil, "there always must be some leaf node, not found for key %s", key)
	leafNode.insertSorted(key, value)
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
	if b.rebalanceCounter != nil {
		b.rebalanceCounter()
	}
	parent := childToRemove.getParent()
	if parent == nil {
		newParent := &innerNode[K, V]{
			children:      []node[K, V]{left, right},
			keys:          []K{separator},
			accessCounter: b.accessCounter,
		}
		left.setParent(newParent)
		right.setParent(newParent)
		return newParent
	}
	assert(!parent.isOverflow(b.order), "parent must not be overflow at this point")
	parent.expandAtChild(childToRemove, left, right, separator)
	left.setParent(parent)
	right.setParent(parent)
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
	keys          []K
	parent        *innerNode[K, V]
	accessCounter accessCounter
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
	n.countAccess()
	foundNodeIndex := len(n.keys) // if no key found, use the last range
	for i, separator := range n.keys {
		if separator > seekedKey {
			foundNodeIndex = i
			break
		}
	}
	// Reached the last range.
	assert(foundNodeIndex < len(n.children), "found node index is outside children range")
	return n.children[foundNodeIndex].findLeafNodeByKey(seekedKey)
}

func (n *innerNode[K, V]) isRoot() bool {
	n.countAccess()
	return n.parent == nil
}

func (n *innerNode[K, V]) isOverflow(order int) bool {
	n.countAccess()
	assert(len(n.children) <= order+1, "there should be no path that results in child len > one more than order, len(children)=%d, order=%d", len(n.children), order)
	return len(n.children) > order
}

func (n *innerNode[K, V]) expandAtChild(childToRemove, left, right node[K, V], separator K) {
	n.countAccess()
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
	n.countAccess()
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
	n.countAccess()
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
	n.countAccess()
	assert(slices.IsSorted(n.keys), "expected keys to be sorted, was: %v", n.keys)
	iMedian := len(n.keys) / 2
	medianValue := n.keys[iMedian]
	leftChildren := slices.Clone(n.children[:iMedian+1]) // clone to allow GC collecting n.children
	leftKeys := slices.Clone(n.keys[:iMedian])
	rightChildren := slices.Clone(n.children[iMedian+1:])
	rightKeys := slices.Clone(n.keys[iMedian+1:])
	newLeft := &innerNode[K, V]{
		children:      leftChildren,
		keys:          leftKeys,
		accessCounter: n.accessCounter,
	}
	for _, c := range leftChildren {
		c.setParent(newLeft)
	}
	newRight := &innerNode[K, V]{
		children:      rightChildren,
		keys:          rightKeys,
		accessCounter: n.accessCounter,
	}
	for _, c := range rightChildren {
		c.setParent(newRight)
	}
	return newLeft, newRight, medianValue
}

func (n *innerNode[K, V]) getParent() *innerNode[K, V] {
	n.countAccess()
	return n.parent
}

func (n *innerNode[K, V]) setParent(p *innerNode[K, V]) {
	n.countAccess()
	n.parent = p
}

func (n *innerNode[K, V]) countAccess() {
	n.accessCounter(n)
}

////////////////////////////////////////
// Leaf node functions and methods
////////////////////////////////////////

// leafNode contains no children, but arbitrary values stored under keys.
type leafNode[K cmp.Ordered, V any] struct {
	pairs         []pair[K, V]
	parent        *innerNode[K, V]
	accessCounter accessCounter
}

type pair[K any, V any] struct {
	key   K
	value V
}

func newLeafNode[K cmp.Ordered, V any](ac accessCounter) *leafNode[K, V] {
	return &leafNode[K, V]{
		pairs:         []pair[K, V]{},
		accessCounter: ac,
	}
}

func (n *leafNode[K, V]) findLeafNodeByKey(seekedKey K) *leafNode[K, V] {
	n.countAccess()
	return n
}

func (n *leafNode[K, V]) getValue(key K) (V, bool) {
	n.countAccess()
	pairs := pairSlice[K, V](n.pairs)
	assert(pairs.isSorted(), "expected pairs to be sorted")
	if i := pairs.bisect(key); i == -1 || n.pairs[i].key != key {
		var zero V
		return zero, false
	} else {
		return n.pairs[i].value, true
	}
}

func (n *leafNode[K, V]) isRoot() bool {
	n.countAccess()
	return n.parent == nil
}

func (n *leafNode[K, V]) isOverflow(order int) bool {
	n.countAccess()
	return len(n.pairs) > order
}

func (n *leafNode[K, V]) getParent() *innerNode[K, V] {
	n.countAccess()
	return n.parent
}

func (n *leafNode[K, V]) setParent(p *innerNode[K, V]) {
	n.countAccess()
	n.parent = p
}

// forceAppend adds key and value regardless if this causes overflow or not.
func (n *leafNode[K, V]) insertSorted(key K, value V) {
	n.countAccess()
	pairs := pairSlice[K, V](n.pairs)
	assert(pairs.isSorted(), "pairs should be sorted before insert")
	i := pairs.bisect(key)
	newPair := pair[K, V]{key: key, value: value}
	if i == -1 {
		n.pairs = append(n.pairs, newPair)
	} else {
		n.pairs = slices.Insert(n.pairs, i, newPair)
	}
	assert(pairSlice[K, V](n.pairs).isSorted(), "pairs should be sorted after insert")
}

func (n *leafNode[K, V]) splitAroundMedian() (*leafNode[K, V], *leafNode[K, V], K) {
	n.countAccess()
	median := n.medianKey()
	left, right := newLeafNode[K, V](n.accessCounter), newLeafNode[K, V](n.accessCounter)
	insertToLeftOrRight := func(p pair[K, V]) {
		if p.key < median {
			left.pairs = append(left.pairs, p)
		} else {
			right.pairs = append(right.pairs, p)
		}
	}
	for _, p := range n.pairs {
		insertToLeftOrRight(p)
	}
	assert(pairSlice[K, V](left.pairs).isSorted(), "left should be sorted")
	assert(pairSlice[K, V](right.pairs).isSorted(), "left should be sorted")
	return left, right, median
}

func (n *leafNode[K, V]) medianKey() K {
	n.countAccess()
	assert(pairSlice[K, V](n.pairs).isSorted(), "expecetd keys to be sorted")
	return n.pairs[len(n.pairs)/2].key
}

func (n *leafNode[K, V]) runRecursiveUntilError(level int, fun func(level int, n node[K, V]) error) error {
	n.countAccess()
	if err := fun(level, n); err != nil {
		return err
	}
	return nil
}

func (n *leafNode[K, V]) print(w io.Writer, indent int) {
	n.countAccess()
	spaces := strings.Repeat(" ", indent)
	for _, p := range n.pairs {
		fmt.Fprintf(w, "%s[%v]:%v\n", spaces, p.key, p.value)
	}
}

func (n *leafNode[K, V]) countAccess() {
	n.accessCounter(n)
}

type pairSlice[K cmp.Ordered, V any] []pair[K, V]

func (s pairSlice[K, V]) isSorted() bool {
	if len(s) == 0 {
		return true
	}
	prev := s[0].key
	for _, p := range s {
		if p.key < prev {
			return false
		}
		prev = p.key
	}
	return true
}

// bisect returns index of the key equal to seeked key or the first larger than seeked key.
func (s pairSlice[K, V]) bisect(key K) int {
	i := sort.Search(len(s), func(i int) bool {
		return s[i].key >= key
	})
	if i == len(s) {
		return -1
	}
	return i
}

// accessCounter is used to inform that a particular node was accessed for sake of profiling.
type accessCounter func(n any)

// rebalanceCounter counts number of re-balances of the nodes.
type rebalanceCounter func()

func dummyAccessCounter(n any) {}
