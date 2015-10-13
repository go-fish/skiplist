package skiplist

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

var (
	listN    int
	number   int
	list     [][]byte
	skipList *SkipList
	readLM   *lockMap
	readM    map[interface{}]interface{}
)

func init() {
	MAXPROCS := runtime.NumCPU()
	runtime.GOMAXPROCS(MAXPROCS)
	listN = MAXPROCS * 10
	number = 100000
	fmt.Println("MAXPROCS is ", MAXPROCS, ", listN is", listN, ", n is ", number, "\n")

	list = make([][]byte, listN, listN)
	for i := 0; i < listN; i++ {
		list1 := make([]byte, 0, number)
		for j := 0; j < number; j++ {
			list1 = append(list1, []byte(strconv.Itoa(j+(i)*number/10))...)
		}
		list[i] = list1
	}

	skipList = NewSkipList()
	readLM = newLockMap()
}

type lockMap struct {
	m  map[interface{}]interface{}
	rw *sync.RWMutex
}

func newLockMap() *lockMap {
	return &lockMap{make(map[interface{}]interface{}), new(sync.RWMutex)}
}

func (t *lockMap) put(k interface{}, v interface{}) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.m[k] = v
}

func (t *lockMap) get(k interface{}) (v interface{}, ok bool) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	v, ok = t.m[k]
	return
}

func BenchmarkLockMapPut(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					readLM.put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMapPut(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[j] = j
			}
			//wg.Done()
		}
	}
}

func BenchmarkSkipListPut(b *testing.B) {
	for n := 0; n < b.N; n++ {

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					skipList.Put([]byte{j}, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkLockMapGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				//itr := NewMapIterator(cm)
				//for itr.HasNext() {
				//	entry := itr.NextEntry()
				//	k := entry.key.(string)
				//	v := entry.value.(int)
				for _, j := range list[k] {
					_, _ = readLM.get(j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMapGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for k := range list[0] {
				_, _ = readM[k]
			}
			//wg.Done()
		}
	}
}

func BenchmarkSkipListGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					_, _ = skipList.Get([]byte{j})
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
