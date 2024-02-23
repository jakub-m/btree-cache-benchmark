package btree

import (
	"cmp"
	"slices"
)

type Btree[KeyT cmp.Ordered, ValueT any] struct {
	// The maximum number of child nodes of a node.
	order int
	// either innerNode or leafNode
	root node[KeyT, ValueT]
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
type node[KeyT cmp.Ordered, ValueT any] interface {
	// findLeafNodeByKey returns the leaf node that holds the value with seeked key, or the one that should
	// hold such a value if it doesn't.
	findLeafNodeByKey(key KeyT) *leafNode[KeyT, ValueT]
}

////////////////////////////////////////
// Btree functions and methods
////////////////////////////////////////

func New[KeyT ~int, ValueT any](order int) *Btree[KeyT, ValueT] {
	root := newLeafNode[KeyT, ValueT]()
	return &Btree[KeyT, ValueT]{
		order: order,
		root:  root,
	}
}

func (b *Btree[KeyT, ValueT]) Find(key KeyT) (ValueT, bool) {
	if n := b.root.findLeafNodeByKey(key); n != nil {
		value, ok := n.values[key]
		return value, ok
	}
	var zero ValueT
	return zero, false
}

func (b *Btree[KeyT, ValueT]) Insert(key KeyT, value ValueT) {
	leafNode := b.root.findLeafNodeByKey(key)
	assert(leafNode != nil, "there always must be some leaf node, not found for key %s", key)

	// https://en.wikipedia.org/wiki/B-tree#Insertion
	if !leafNode.isFull(b.order) {
		leafNode.insertAssumingHasSpace(key, value)
	}
	// The leaf node is full, so need to split.
	medianKey := leafNode.medianKeyForChildrenAndKey(key)
	leftLeaf, rightLeaf := leafNode.splitAroundMedian(medianKey)

}

////////////////////////////////////////
// Inner node functions and methods
////////////////////////////////////////

// innerNode has children nodes that are either innerNodes or leafNodes.
type innerNode[KeyT cmp.Ordered, ValueT any] struct {
	children []node[KeyT, ValueT]
	// keys separate children. For m children there is always m-1 keys.
	keys   []KeyT
	parent *innerNode[KeyT, ValueT]
}

func (n *innerNode[KeyT, ValueT]) findLeafNodeByKey(seekedKey KeyT) *leafNode[KeyT, ValueT] {
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

////////////////////////////////////////
// Leaf node functions and methods
////////////////////////////////////////

// leafNode contains no children, but arbitrary values stored under keys.
type leafNode[KeyT cmp.Ordered, ValueT any] struct {
	values map[KeyT]ValueT
	parent *innerNode[KeyT, ValueT]
}

func newLeafNode[KeyT cmp.Ordered, ValueT any]() *leafNode[KeyT, ValueT] {
	return &leafNode[KeyT, ValueT]{
		values: make(map[KeyT]ValueT),
	}
}

func (n *leafNode[KeyT, ValueT]) findLeafNodeByKey(seekedKey KeyT) *leafNode[KeyT, ValueT] {
	return n
}

func (n *leafNode[KeyT, ValueT]) insertAssumingHasSpace(key KeyT, value ValueT) {
	n.values[key] = value
}

func (n *leafNode[KeyT, ValueT]) isFull(order int) bool {
	assert(len(n.values) < order, "length of values array must be smaller than order")
	return len(n.values) == (order - 1)
}

// medianKeyForChildrenAndKey return the median out of children elements and the new key.
func (n *leafNode[KeyT, ValueT]) medianKeyForChildrenAndKey(key KeyT) KeyT {
	keys := make([]KeyT, 0, len(n.values)+1)
	for k := range n.values {
		keys = append(keys, k)
	}
	keys[len(n.values)] = key // insert last key
	slices.Sort(keys)
	return keys[len(keys)/2]
}

func (n *leafNode[KeyT, ValueT]) splitAroundMedian(key KeyT) (*leafNode[KeyT, ValueT], *leafNode[KeyT, ValueT]) {
	left := newLeafNode[KeyT, ValueT]()
}

/// func (n *node[KeyT, ValueT]) createLeftNodeAfterSplit(median KeyT) *node[KeyT, ValueT] {
/// 	leftChildren := make([]any, len(n.children))
/// 	rightChildren := make([]any, len(n.children))
/// 	if n.isLeaf() {
/// 		for _, c := range n.children {
/// 			kv := c.(*keyValue[KeyT, ValueT])
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
/// 	// return &node[KeyT, ValueT]{
/// 	// 	children: ...,
/// 	// 	childrenCount, ...
/// 	// 	keys: ...
/// 	// }
/// }

///////////////////////////////////////////////////////

/// // insertChildInOrder assumes there is space and there is no inserted item with `key`.
/// func (n *node[KeyT, ValueT]) insertChildInOrder(key KeyT, value ValueT) {
/// 	assert(n.isLeaf(), "assumed leaf node")
/// 	assert(n.childrenCount < len(n.children), "there is no spare capacity left in the array, insertChildInOrder should not be run at all.")
/// 	insertAtIndex := 0
/// 	// Find the index at which to insert the new key-value.
/// 	for _, child := range n.children {
/// 		if child == nil {
/// 			break
/// 		}
/// 		childKv := child.(*keyValue[KeyT, ValueT])
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
/// 	n.children[insertAtIndex] = &keyValue[KeyT, ValueT]{key: key, value: &value}
/// 	n.childrenCount++
/// 	for i := insertAtIndex + 1; i < len(n.children); i++ {
/// 		next = n.children[i]
/// 		n.children[i] = curr
/// 		curr = next
/// 	}
/// }

/// func (b *Btree[KeyT, ValueT]) ValidityCheck() error {
/// 	check := func(n *node[KeyT, ValueT]) error {
/// 		if n.isLeaf() {
/// 			var prevKey *KeyT
/// 			for _, child := range n.children {
/// 				kv := child.(*keyValue[KeyT, ValueT])
/// 				if prevKey != nil {
/// 					if !(*prevKey < kv.key) {
/// 						return fmt.Errorf("for a child, prev key=%d, next key=%d", *prevKey, kv.key)
/// 					}
/// 				}
/// 				prevKey = &kv.key
/// 			}
/// 		}
/// 		return nil
/// 	}
/// 	return b.root.runRecursiveUntilError(check)
/// }

/// func (n *node[KeyT, ValueT]) runRecursiveUntilError(fun func(n *node[KeyT, ValueT]) error) error {
/// 	if err := fun(n); err != nil {
/// 		return err
/// 	}
/// 	for _, child := range n.children {
/// 		if child != nil {
/// 			if n2, ok := child.(*node[KeyT, ValueT]); ok {
/// 				if err := n2.runRecursiveUntilError(fun); err != nil {
/// 					return err
/// 				}
/// 			}
/// 		}
/// 	}
/// 	return nil
/// }
///
/// func (b *Btree[KeyT, ValueT]) Print(w io.Writer) {
/// 	b.root.print(w, 0)
/// }
///
/// func (n *node[KeyT, ValueT]) print(w io.Writer, indent int) {
/// 	spaces := strings.Repeat(" ", indent)
/// 	fmt.Fprintf(w, "%schildren: %d isLeaf: %t\n", spaces, len(n.children), n.isLeaf())
/// 	if n.isLeaf() {
/// 		for i, child := range n.children {
/// 			if child == nil {
/// 				fmt.Fprintf(w, "%s[%d:nil] nil\n", spaces, i)
/// 			} else {
/// 				kv := child.(*keyValue[KeyT, ValueT])
/// 				fmt.Fprintf(w, "%s[%d:%d] %s\n", spaces, i, kv.key, fmt.Sprint(kv.value))
/// 			}
/// 		}
/// 	} else {
/// 		for i, key := range n.keys {
/// 			fmt.Fprintf(w, "%s%d", spaces, key)
/// 			if i < (len(n.keys) - 1) {
/// 				fmt.Fprintf(w, " | ")
/// 			}
/// 			fmt.Fprintf(w, "\n")
/// 			for _, child := range n.children {
/// 				n2 := child.(*node[KeyT, ValueT])
/// 				n2.print(w, indent+1)
/// 			}
/// 		}
/// 		for _, child := range n.children {
/// 			kv := child.(*keyValue[KeyT, ValueT])
/// 			fmt.Fprintf(w, "%s[k:%d] %s\n", spaces, kv.key, fmt.Sprint(kv.value))
/// 		}
/// 	}
/// }
///
