package skiplist

import (
	"bytes"
	"errors"
	"unsafe"
)

type SubMap struct {
	//Underlying Skiplist
	skipList *SkipList

	//lower bound key, or nil if from start
	lo []byte

	//upper bound key, or nil if to end
	hi []byte

	//direction
	isDescending bool
}

func NewSubMap(skipList *SkipList, fromKey, toKey []byte, isDescending bool) (*SubMap, error) {
	if fromKey != nil && toKey != nil && bytes.Compare(fromKey, toKey) > 0 {
		return nil, errors.New("Inconsistent Range")
	}

	return &SubMap{
		skipList:     skipList,
		lo:           fromKey,
		hi:           toKey,
		isDescending: isDescending,
	}, nil
}

func (this *SubMap) tooLow(key []byte) bool {
	if this.lo != nil {
		if bytes.Compare(key, this.lo) < 0 {
			return true
		}
	}

	return false
}

func (this *SubMap) tooHigh(key []byte) bool {
	if this.hi != nil {
		if bytes.Compare(key, this.hi) > 0 {
			return true
		}
	}

	return false
}

func (this *SubMap) isBounds(key []byte) bool {
	return !this.tooLow(key) && !this.tooHigh(key)
}

func (this *SubMap) checkKeyBounds(key []byte) error {
	if key == nil {
		return ErrNilKey
	}

	if !this.isBounds(key) {
		return errors.New("Key Out Of Range")
	}

	return nil
}

//Returns true if node key is less than upper bound of range
func (this *SubMap) isBeforeEnd(n *Node) bool {
	if n == nil {
		return false
	}

	if this.hi == nil {
		return true
	}

	if n.Key == nil {
		return true
	}

	if bytes.Compare(n.Key, this.hi) > 0 {
		return false
	}

	return true
}

//Returns lowest node. This node might not be in range, so most usages need to check bounds
func (this *SubMap) loNode() *Node {
	if this.lo == nil {
		return this.skipList.findFirst()
	} else {
		return this.skipList.findNear(this.lo, GT|EQ)
	}
}

//Returns highest node. This node might not be in range, so most usages need to check bounds
func (this *SubMap) hiNode() *Node {
	if this.hi == nil {
		return this.skipList.findLast()
	} else {
		return this.skipList.findNear(this.hi, LT|EQ)
	}
}

type Iterator struct {
	//the last node returned by next()
	LastReturned *Node

	//the next node to return from next()
	Next *Node

	//Cache of next value field to maintain weak consistency
	NextValue interface{}

	subMap *SubMap
}

func CreateIteratorFromSubMap(subMap *SubMap) *Iterator {
	var this = &Iterator{
		subMap: subMap,
	}
	for {
		if subMap.isDescending {
			this.Next = subMap.hiNode()
		} else {
			this.Next = subMap.loNode()
		}

		if this.Next == nil {
			break
		}

		var x = this.Next.Value
		if x != nil && x != unsafe.Pointer(this.Next) {
			if !subMap.isBounds(this.Next.Key) {
				this.Next = nil
			} else {
				this.NextValue = *((*interface{})(x))
			}

			break
		}
	}

	return this
}

func (this *Iterator) HasNext() bool {
	return this.Next != nil
}

func (this *Iterator) advance() error {
	if this.Next == nil {
		return errors.New("No Next Node")
	}

	this.LastReturned = this.Next
	if this.subMap.isDescending {
		this.descend()
	} else {
		this.ascend()
	}

	return nil
}

func (this *Iterator) ascend() {
	for {
		this.Next = (*Node)(this.Next.Next)
		if this.Next == nil {
			break
		}

		var x = this.Next.Value
		if x != nil && x != unsafe.Pointer(this.Next) {
			if this.subMap.tooHigh(this.Next.Key) {
				this.Next = nil
			} else {
				this.NextValue = *((*interface{})(x))
			}

			break
		}
	}
}

func (this *Iterator) descend() {
	for {
		this.Next = this.subMap.skipList.findNear(this.LastReturned.Key, LT)
		if this.Next == nil {
			break
		}

		var x = this.Next.Value
		if x != nil && x != unsafe.Pointer(this.Next) {
			if this.subMap.tooLow(this.Next.Key) {
				this.Next = nil
			} else {
				this.NextValue = *((*interface{})(x))
			}

			break
		}
	}
}

func (this *Iterator) Remove() error {
	var l = this.LastReturned
	if l == nil {
		return errors.New("Illegal Last Return Node")
	}

	this.subMap.skipList.Remove(l.Key)
	this.LastReturned = nil

	return nil
}

func (this *Iterator) NextNode() ([]byte, interface{}, error) {
	var n = this.Next
	var v = this.NextValue

	var err = this.advance()
	if err != nil {
		return nil, nil, err
	}

	return n.Key, v, nil
}
