package skiplist

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	key    []byte
	value  unsafe.Pointer
	next   *node
	marked bool
}

//create new node
func newNode(key []byte, value unsafe.Pointer, next *node, prev *node) *node {
	return &node{
		key:   key,
		value: value,
		next:  next,
	}
}

//delete marked node
func (n *node) deleteMarkedNode(prev, succ *node) {
	if n == prev.next && succ == n.next {
		prev.casNext(n, succ)
	}
}

//cas next
func (n *node) casNext(cmp, succ *node) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next)), unsafe.Pointer(cmp), unsafe.Pointer(succ))
}
