# skiplist
concurrent skiplist transplant from java

<h3>Current Version: 1.0 beta</h3>

<h5>Put, Get, Remove, SubMap, Iterator</h5>

```Go
var skipList = skiplist.NewSkipList()

//put return the prev value and error
previou, err := skiplist.Put([]byte("test"), 10) //return nil, nil

//put only if the key does not exists
previou, err = skiplist.PutOnlyIfAbsent([]byte("test"), 100) //return 10, nil

//get
res, err := skiplist.Get([]byte("test")) //return 10, nil

//update
previou, err = skiplist.Update([]byte("test"), func(oldValue interface{}) interface{}{
	return 2 * oldValue.(int)
}) //return 10, nil

//update only if the key does not exists
previou, err = skiplist.UpdateOnlyIfAbsent([]byte("test"), func(oldValue interface{}) interface{}{
	return 2 * oldValue.(int)
}) //return 20, nil

//remove only if values eq
var ok bool
previou, ok, err = skiplist.CompareAndRemove([]byte("test"), 100) //return 20, false, nil

//remove
previou, ok, err = skiplist.Remove([]byte("test")) //return 20, true, nil
```

