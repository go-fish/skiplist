package skiplist

import (
	"sync/atomic"
	"unsafe"
)

type Index struct {
	Node *Node

	Down, Right unsafe.Pointer

	Level int
}

func createIndex(node *Node, down, right *Index) *Index {
	return &Index{
		Node:  node,
		Down:  unsafe.Pointer(down),
		Right: unsafe.Pointer(right),
	}
}

func createHeadIndex(node *Node, down, right *Index, level int) *Index {
	var index = createIndex(node, down, right)
	index.Level = level

	return index
}

func (this *Index) casRight(cmp, val unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&this.Right, cmp, val)
}

//@return true if indexed node is known to be deleted
func (this *Index) indexesDeletedNode() bool {
	return this.Node.Value == nil
}

func (this *Index) link(succ, newSucc *Index) bool {
	var n = this.Node
	newSucc.Right = unsafe.Pointer(succ)

	return n.Value != nil && this.casRight(unsafe.Pointer(succ), unsafe.Pointer(newSucc))
}

func (this *Index) unlink(succ *Index) bool {
	return !this.indexesDeletedNode() && this.casRight(unsafe.Pointer(succ), succ.Right)
}
