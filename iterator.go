package skiplist

import "unsafe"

type Iterator struct {
	sl       *Skiplist
	lo       []byte
	hi       []byte
	next     *node
	lastNode *node
}

//create iterator
func NewIterator(sl *Skiplist, fromKey, toKey []byte) (*Iterator, error) {
	i := &Iterator{
		sl: sl,
		lo: fromKey,
		hi: toKey,
	}

	if fromKey == nil {
		i.next = sl.findFirstNode()
		return i, nil
	}

	var exactMatch bool
	if i.next, exactMatch = sl.findPrecursorOrNode(fromKey); !exactMatch {
		return nil, ErrUnknownFromKey
	}

	return i, nil
}

//next
func (i *Iterator) Next() bool {
	return i.next != nil
}

//get next node
func (i *Iterator) NextNode() ([]byte, unsafe.Pointer) {
	i.lastNode = i.next
	i.next = i.next.next

	return i.lastNode.key, i.lastNode.value
}

//remove
func (i *Iterator) Remove() error {
	if i.next == nil {
		return ErrRemoveNilNode
	}

	i.sl.remove(i.lastNode.key, i.lastNode.value)
	return nil
}
