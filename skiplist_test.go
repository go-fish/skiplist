package skiplist

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNil(t *testing.T) {
	convey.Convey("Nil Can Not Be Key Or Value/Action", t, func() {
		skiplist := NewSkipList()
		_, err := skiplist.Put(nil, 111)
		convey.So(err, convey.ShouldNotBeNil)

		var nilKey []byte = nil
		_, err = skiplist.Put(nilKey, 111)
		convey.So(err, convey.ShouldNotBeNil)

		_, err = skiplist.Put([]byte{'1'}, nil)
		convey.So(err, convey.ShouldNotBeNil)

		_, err = skiplist.Get(nil)
		convey.So(err, convey.ShouldNotBeNil)

		_, err = skiplist.Get(nilKey)
		convey.So(err, convey.ShouldNotBeNil)

		_, _, err = skiplist.Remove(nil)
		convey.So(err, convey.ShouldNotBeNil)

		_, _, err = skiplist.Remove(nilKey)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestSkipList(t *testing.T) {
	var key []byte = []byte{'t', 'e', 's', 't'}
	var value interface{} = 111

	skiplist := NewSkipList()

	//test put
	previou, err := skiplist.Put(key, value)
	if previou != nil || err != nil {
		t.Errorf("Put %v, %v , return %v, %v, want nil, nil", key, value, previou, err)
	}

	//test get key
	res, err := skiplist.Get(key)
	if res != value || err != nil {
		t.Errorf("Get %v , return %v, %v, want %v, nil", key, res, err, value)
	}

	//test put same key
	v := rand.Float64()
	previou, err = skiplist.Put(key, v)
	if previou != value || err != nil {
		t.Errorf("Put %v, %v second time, return %v, %v, want %v, nil", key, v, previou, err, value)
	}

	//test get key
	res, err = skiplist.Get(key)
	if res.(float64) != v || err != nil {
		t.Errorf("Get %v , return %v, %v, want %v, nil", key, res.(float64), err, v)
	}

	//test put only if absent, only if key is not exists, put the value to skiplist
	previou, err = skiplist.PutOnlyIfAbsent(key, value)
	if previou.(float64) != v || err != nil {
		t.Errorf("Put %v, %v, return %v, %v, want %v, nil", key, value, previou, err, v)
	}

	//test get key
	res, err = skiplist.Get(key)
	if res.(float64) != v || err != nil {
		t.Errorf("Get %v , return %v, %v, want %v, nil", key, res.(float64), err, v)
	}

	var secondKey []byte = []byte{'t', 'e'}
	previou, err = skiplist.PutOnlyIfAbsent(secondKey, value)
	if previou != nil || err != nil {
		t.Errorf("Put %v, %v , return %v, %v, want nil, nil", secondKey, value, previou, err)
	}

	//test get key
	res, err = skiplist.Get(secondKey)
	if res != value || err != nil {
		t.Errorf("Get %v , return %v, %v, want %v, nil", secondKey, res, err, value)
	}

	//test remove key only if value of secondKey == v
	ok, err := skiplist.CompareAndRemove(secondKey, v)
	if ok != false || err != nil {
		t.Errorf("Remove %v, %v , return %v, %v, want false, nil", secondKey, v, ok, err)
	}

	//test get key
	res, err = skiplist.Get(secondKey)
	if res != value || err != nil {
		t.Errorf("Get %v , return %v, %v, want %v, nil", secondKey, res, err, value)
	}

	//test remove key
	previou, ok, err = skiplist.Remove(secondKey)
	if previou != value || ok != true || err != nil {
		t.Errorf("Remove %v , return %v, %v, %v, want %v, true, nil", key, previou, ok, err, value)
	}

	//test get key
	res, err = skiplist.Get(secondKey)
	if res != nil || err != nil {
		t.Errorf("Get %v , return %v, %v, want nil, nil", secondKey, res, err)
	}

	//test update key
	previou, err = skiplist.Update(key, func(oldValue interface{}) interface{} {
		return float64(2) * oldValue.(float64)
	})
	if previou != v || err != nil {
		t.Errorf("Update %v , return %v, %v, want %v, nil", key, previou, err, v)
	}

	//test get key
	res, err = skiplist.Get(key)
	if res != float64(2)*v || err != nil {
		t.Errorf("Get %v , return %v, %v, want %v, nil", key, res, err, float64(2)*v)
	}

	var key1 = []byte{'t', 'e', 's', 't', '1', '1', '1'}
	skiplist.Put(key1, 111)
	skiplist.Put([]byte{'t', 'e', 's', 't', '1', '1', '2'}, 112)
	skiplist.Put([]byte{'t', 'e', 's', 't', '1', '1', '3'}, 113)

	//test size
	if skiplist.Size != 4 {
		t.Errorf("Fail To Get Size Info, return %d, want 4", skiplist.Size)
	}

	//test iterator
	submap, _ := NewSubMap(skiplist, nil, nil, false)

	iterator := CreateIteratorFromSubMap(submap)

	if iterator.HasNext() != true {
		t.Errorf("Fail To Get Next Info, return false, want true")
	}

	k, val, err := iterator.NextNode()
	if bytes.Compare(k, key) != 0 || val != float64(2)*v || err != nil {
		t.Errorf("Get Next Node, return %v, %v, %v, want %v, %v, nil", k, val, err, key, float64(2)*v)
	}

	if iterator.HasNext() != true {
		t.Errorf("Fail To Get Next Info, return false, want true")
	}
	//
	//	k, val, err = iterator.NextNode()
	//	if bytes.Compare(k, []byte{'t', 'e', 's', 't', '1', '1', '1'}) != 0 || val != 111 || err != nil {
	//		t.Errorf("Get Next Node, return %v, %v, %v, want test111, 111, nil", k, val, err)
	//	}
}
