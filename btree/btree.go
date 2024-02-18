package btree

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
		children: make([]any, 0, order),
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
		panic("todo")
	} else {
		leaf.insertChildInOrder(key, value)
	}
}

// insertChildInOrder assumes there is space and there is no inserted item with `key`.
func (n *node[KeyT, ValueT]) insertChildInOrder(key KeyT, value ValueT) {
	assert(n.isLeaf(), "assumed leaf node")
	for i, child := range n.children {
		childKv := child.(*keyValue[KeyT, ValueT])
		assert(childKv.key != key, "the case that the equal key is in the tree should have been already handled: key=%d", key)
		if childKv.key > key {
			// Here is time to insert KV. Move all the values to the right. The capacity should be already there.
			var curr, next *keyValue[KeyT, ValueT]
			curr = &keyValue[KeyT, ValueT]{key: key, value: &value}
			for j := i; j < len(n.children)+1; j++ {
				next = n.children[j].(*keyValue[KeyT, ValueT])
				n.children[j] = curr
				curr = next
			}
		}
	}
}

// isLeaf says if the node is a leaf node, that is, a node that's children are the pointers to the actually
// stored values.
func (n *node[KeyT, ValueT]) isLeaf() bool {
	return n.keys == nil
}

func (b *Btree[KeyT, ValueT]) isNodeFull(n *node[KeyT, ValueT]) bool {
	assert(len(n.children) <= b.order, "too much child nodes")
	return len(n.children) == b.order
}

func (b *Btree[KeyT, ValueT]) ValidityCheck() error {
	return nil
}
