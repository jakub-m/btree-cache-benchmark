package btree

type KeyT int

type Btree struct {
	// The maximum number of child nodes of a node.
	order int
	root  *node
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
type node struct {
	// children nodes, either internal nodes, or the actually inserted items (keyValue structures).
	children []any
	// keys separte children. There are len(children) - 1 keys at any moments.
	// Leaf nodes (the nodes that contain the the actually inserted items) do not have keys, the keys
	// array is set to nil.
	keys []KeyT
}

// keyValue is the struct that is
type keyValue struct {
	key   KeyT
	value any
}

func New(order int) *Btree {
	root := node{
		children: make([]any, 0, order),
		keys:     nil,
	}
	return &Btree{
		order: order,
		root:  &root,
	}
}

func (b *Btree) Find(key KeyT) any {
	if _, keyValue := b.root.findLeafNodeByKey(key); keyValue == nil {
		return nil
	} else {
		return keyValue.value
	}
}

// findLeafNodeByKey returns the leaf node that holds the value with seeked key, or the one that should
// hold such a value if it doesn't.
func (n *node) findLeafNodeByKey(seekedKey KeyT) (*node, *keyValue) {
	if n.isLeaf() {
		for _, c := range n.children {
			kv := c.(*keyValue)
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
				child := n.children[i].(*node)
				return child.findLeafNodeByKey(seekedKey)
			}
		}
		// Reached the last range.
		child := n.children[len(n.children)-1].(*node)
		return child.findLeafNodeByKey(seekedKey)
	}
}

// isLeaf says if the node is a leaf node, that is, a node that's children are the pointers to the actually
// stored values.
func (n *node) isLeaf() bool {
	return n.keys == nil
}

func (b *Btree) Insert[T ~KeyT](key T, value any) {

}
