package skiplist

import (
	"sync/atomic"
	"unsafe"
)

type Node struct {
	Key   []byte
	Value unsafe.Pointer

	Next unsafe.Pointer
}

func createNode(key []byte, value unsafe.Pointer, next unsafe.Pointer) *Node {
	return &Node{
		Key:   key,
		Value: value,
		Next:  next,
	}
}

func createMarker(next unsafe.Pointer) *Node {
	var node = &Node{}

	node.Value = unsafe.Pointer(node)
	node.Next = next

	return node
}

func (this *Node) casValue(cmp, val unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&this.Value, cmp, val)
}

func (this *Node) casNext(cmp, val unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&this.Next, cmp, val)
}

func (this *Node) isMarker() bool {
	return this.Value == unsafe.Pointer(this)
}

var BaseHeader = 0

func (this *Node) isBaseHeader() bool {
	return this.Value == unsafe.Pointer(&BaseHeader)
}

func (this *Node) appendMarker(f *Node) bool {
	return this.casNext(unsafe.Pointer(f), unsafe.Pointer(createMarker(unsafe.Pointer(f))))
}

func (this *Node) helpDelete(b, f *Node) {
	var f1 = unsafe.Pointer(f)
	var t = unsafe.Pointer(this)

	if f1 == this.Next && t == b.Next {
		if f == nil || f.Value != f1 {
			this.casNext(f1, unsafe.Pointer(createMarker(f1)))
		} else {
			b.casNext(t, f.Next)
		}
	}
}
