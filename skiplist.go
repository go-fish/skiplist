package skiplist

import (
	"bytes"
	"errors"
	"math"
	"math/rand"
	"sync/atomic"
	"unsafe"
)

var (
	ErrNilKey   = errors.New("Unsupport Nil Key")
	ErrNilValue = errors.New("Unsupport Nil Value")
)

type SkipList struct {
	Head unsafe.Pointer

	MaxLevel int

	Ip int

	Size int32
}

func NewSkipList() *SkipList {
	return &SkipList{
		Head:     unsafe.Pointer(createHeadIndex(createNode(nil, unsafe.Pointer(&BaseHeader), nil), nil, nil, 1)),
		MaxLevel: 32,
		Ip:       int(math.Ceil(1 / DefaultProbability)),
	}
}

func (this *SkipList) findPredecessor(key []byte) *Node {
	for {
		var q = (*Index)(this.Head)
		var r = (*Index)(q.Right)

		for {
			if r != nil {
				var n = r.Node
				var k = n.Key

				if n.Value == nil {
					if !q.unlink(r) {
						break
					}

					r = (*Index)(q.Right)
					continue
				}

				if bytes.Compare(key, k) > 0 {
					q = r
					r = (*Index)(r.Right)
					continue
				}
			}

			var d = (*Index)(q.Down)
			if d != nil {
				q = d
				r = (*Index)(d.Right)
			} else {
				return q.Node
			}
		}
	}
}

func (this *SkipList) Put(key []byte, value interface{}) (interface{}, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if value == nil {
		return nil, ErrNilValue
	}

	return this.put(key, value, nil, false), nil
}

func (this *SkipList) PutOnlyIfAbsent(key []byte, value interface{}) (interface{}, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if value == nil {
		return nil, ErrNilValue
	}

	return this.put(key, value, nil, true), nil
}

func (this *SkipList) Update(key []byte, action func(oldValue interface{}) interface{}) (interface{}, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if action == nil {
		return nil, errors.New("Unsupport Nil Action In Func Update")
	}

	return this.put(key, nil, action, false), nil
}

func (this *SkipList) UpdateOnlyIfAbsent(key []byte, action func(oldValue interface{}) interface{}) (interface{}, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	if action == nil {
		return nil, errors.New("Unsupport Nil Action In Func Update")
	}

	return this.put(key, nil, action, true), nil
}

func (this *SkipList) put(key []byte, value interface{}, action func(oldValue interface{}) interface{}, onlyIfAbsent bool) interface{} {
	for {
		var b = this.findPredecessor(key)
		var n = (*Node)(b.Next)

		for {
			if n != nil {
				var f = (*Node)(n.Next)
				if n != (*Node)(b.Next) {
					break
				}

				var v = n.Value
				if v == nil {
					n.helpDelete(b, f)
					break
				}

				if v == unsafe.Pointer(n) || b.Value == nil {
					break
				}

				var c = bytes.Compare(key, n.Key)
				if c > 0 {
					b = n
					n = f
					continue
				}
				if c == 0 {
					//get new value
					var newValue interface{}
					if value != nil {
						newValue = value
					} else {
						newValue = action(*((*interface{})(v)))
					}

					if onlyIfAbsent || n.casValue(v, unsafe.Pointer(&newValue)) {
						if v == nil {
							return nil
						} else {
							return *((*interface{})(v))
						}
					} else {
						break
					}
				}
			}

			//get new value
			var newValue interface{}
			if value != nil {
				newValue = value
			} else {
				newValue = action(nil)
			}

			var z = createNode(key, unsafe.Pointer(&newValue), unsafe.Pointer(n))
			if b.casNext(unsafe.Pointer(n), unsafe.Pointer(z)) {
				atomic.AddInt32(&(this.Size), 1)
			} else {
				break
			}

			var level = this.randomLevel()
			if level > 0 {
				this.insertIndex(z, level)
			}

			return nil
		}
	}
}

func (this *SkipList) Get(key []byte) (interface{}, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	return this.get(key), nil
}

func (this *SkipList) ContainsKey(key []byte) bool {
	return this.get(key) != nil
}

func (this *SkipList) get(key []byte) interface{} {
	for {
		var n = this.findNode(key)
		if n == nil {
			return nil
		}

		if n.Value == nil {
			return nil
		} else {
			return *((*interface{})(n.Value))
		}
	}
}

func (this *SkipList) Remove(key []byte) (interface{}, bool, error) {
	if key == nil {
		return nil, false, ErrNilKey
	}

	return this.remove(key, nil), true, nil
}

func (this *SkipList) CompareAndRemove(key []byte, value interface{}) (bool, error) {
	if key == nil {
		return false, ErrNilKey
	}

	return this.remove(key, value) != nil, nil
}

func (this *SkipList) remove(key []byte, value interface{}) interface{} {
	for {
		var b = this.findPredecessor(key)
		var n = (*Node)(b.Next)

		for {
			if n == nil {
				return nil
			}

			var f = (*Node)(n.Next)
			if n != (*Node)(b.Next) {
				break
			}

			var v = n.Value
			if v == nil {
				n.helpDelete(b, f)
				break
			}

			if v == unsafe.Pointer(n) || b.Value == nil {
				break
			}

			var c = bytes.Compare(key, n.Key)
			if c < 0 {
				return nil
			}
			if c > 0 {
				b = n
				n = f
				continue
			}

			if value != nil && value != *((*interface{})(v)) {
				return nil
			}

			if !n.casValue(n.Value, nil) {
				break
			}

			if !n.appendMarker(f) || !b.casNext(unsafe.Pointer(n), unsafe.Pointer(f)) {
				this.findNode(key)
			} else {
				this.findPredecessor(key)
				var head = (*Index)(this.Head)
				if head.Right == nil {
					this.tryReduceLevel()
				}

				atomic.AddInt32(&(this.Size), -1)
			}

			if v == nil {
				return nil
			} else {
				return *((*interface{})(v))
			}
		}
	}
}

func (this *SkipList) tryReduceLevel() {
	var h = (*Index)(this.Head)
	var d = (*Index)(h.Down)
	var e *Index
	if d != nil {
		e = (*Index)(d.Down)
	}

	if h.Level > 3 && d != nil && e != nil && e.Right == nil && d.Right == nil && h.Right == nil && this.casHead(this.Head, h.Down) && h.Right != nil {
		this.casHead(h.Down, d.Down)
	}
}

const DefaultProbability float64 = 0.25

func (this *SkipList) randomLevel() int {
	var level = 1

	for level < this.MaxLevel && rand.Intn(this.Ip) == 0 {
		level++
	}

	return level
}

func (this *SkipList) insertIndex(z *Node, level int) {
	var h = (*Index)(this.Head)
	var max = h.Level

	if level <= max {
		var idx *Index
		for i := 1; i <= level; i++ {
			idx = createIndex(z, idx, nil)
		}

		this.addIndex(idx, h, level)
	} else {
		level = max + 1
		var idxs = make([]*Index, level+1, level+1)
		var idx *Index

		for i := 1; i <= level; i++ {
			idx = createIndex(z, idx, nil)
			idxs[i] = idx
		}

		var oldh *Index
		var k int
		for {
			oldh = (*Index)(this.Head)
			var oldhLevel = oldh.Level
			if level <= oldhLevel {
				k = level
				break
			}

			var newh = oldh
			var oldNode = oldh.Node
			for j := oldhLevel + 1; j <= level; j++ {
				newh = createHeadIndex(oldNode, newh, idxs[j], j)
			}

			if this.casHead(unsafe.Pointer(oldh), unsafe.Pointer(newh)) {
				k = oldhLevel
				break
			}
		}

		this.addIndex(idxs[k], oldh, k)
	}
}

func (this *SkipList) casHead(cmp, val unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&this.Head, cmp, val)
}

func (this *SkipList) addIndex(idx, h *Index, level int) {
	var insertionLevel = level

	for {
		var j = h.Level
		var q = h
		var r = (*Index)(q.Right)
		var t = idx

		for {
			if r != nil {
				var n = r.Node
				var c = bytes.Compare(idx.Node.Key, n.Key)
				if n.Value == nil {
					if !q.unlink(r) {
						break
					}

					r = (*Index)(q.Right)
					continue
				}

				if c > 0 {
					q = r
					r = (*Index)(r.Right)
					continue
				}
			}

			if j == insertionLevel {
				if t.indexesDeletedNode() {
					this.findNode(idx.Node.Key)
					return
				}

				if !q.link(r, t) {
					break
				}

				insertionLevel--
				if insertionLevel == 0 {
					if t.indexesDeletedNode() {
						this.findNode(idx.Node.Key)
					}
					return
				}
			}

			j--
			if j >= insertionLevel && j < level {
				t = (*Index)(t.Down)
			}

			q = (*Index)(q.Down)
			r = (*Index)(q.Right)
		}
	}
}

func (this *SkipList) findNode(key []byte) *Node {
	for {
		var b = this.findPredecessor(key)
		var n = (*Node)(b.Next)

		for {
			if n == nil {
				return nil
			}

			var f = (*Node)(n.Next)
			if n != (*Node)(b.Next) {
				break
			}

			if n.Value == nil {
				n.helpDelete(b, f)
				break
			}
			if n.Value == unsafe.Pointer(n) || b.Value == nil {
				break
			}

			var c = bytes.Compare(key, n.Key)
			if c == 0 {
				return n
			}
			if c < 0 {
				return nil
			}

			b = n
			n = f
		}
	}
}

func (this *SkipList) findFirst() *Node {
	for {
		var h = (*Index)(this.Head)
		var b = h.Node
		var n = (*Node)(b.Next)

		if n == nil {
			return nil
		}

		if n.Value != nil {
			return n
		}

		n.helpDelete(b, (*Node)(n.Next))
	}
}

func (this *SkipList) findLast() *Node {
	var q = (*Index)(this.Head)
	for {
		var r = (*Index)(q.Right)
		var d = (*Index)(q.Down)
		if r != nil {
			if r.indexesDeletedNode() {
				q.unlink(r)
				q = (*Index)(this.Head)
			} else {
				q = r
			}
		} else if d != nil {
			q = d
		} else {
			var b = q.Node
			var n = (*Node)(b.Next)

			for {
				if n == nil {
					if b.isBaseHeader() {
						return nil
					} else {
						return b
					}
				}

				var f = (*Node)(n.Next)
				if n != (*Node)(b.Next) {
					break
				}

				var v = n.Value
				if v == nil {
					n.helpDelete(b, f)
					break
				}

				if v == unsafe.Pointer(n) || b.Value == nil {
					break
				}

				b = n
				n = f
			}

			q = (*Index)(this.Head)
		}
	}
}

const (
	GT = iota
	EQ
	LT
)

/**
 * Utility for ceiling, floor, lower, higher methods.
 * @param key the key
 * @param rel the relation -- OR'ed combination of EQ, LT, GT
 * @return nearest node fitting relation, or nil if no such
 */
func (this *SkipList) findNear(key []byte, rel int) *Node {
	for {
		var b = this.findPredecessor(key)
		var n = (*Node)(b.Next)

		for {
			if n == nil {
				if (rel&LT) == 0 || b.isBaseHeader() {
					return nil
				} else {
					return b
				}
			}

			var f = (*Node)(n.Next)
			if n != (*Node)(b.Next) {
				break
			}

			var v = n.Value
			if v == nil {
				n.helpDelete(b, f)
				break
			}

			if v == unsafe.Pointer(n) || b.Value == nil {
				break
			}

			var c = bytes.Compare(key, n.Key)
			if (c == 0 && (rel&EQ) != 0) || (c < 0 && (rel&LT == 0)) {
				return n
			}

			if c <= 0 && (rel&LT) != 0 {
				if b.isBaseHeader() {
					return nil
				} else {
					return b
				}
			}

			b = n
			n = f
		}
	}
}
