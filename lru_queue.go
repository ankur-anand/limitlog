package limitlog

import (
	"container/list"
)

type entry struct {
	key int
}

type LRUCache struct {
	size  int                   // capacity
	list  *list.List            // doubly linked list
	items map[int]*list.Element // hash table for list.List.Element existence check
}

// NewLRU constructs an LRU of the given size
func NewLRU(size int) *LRUCache {
	if size <= 0 {
		panic("Must provide a positive size")
	}
	c := &LRUCache{
		size:  size,
		list:  list.New(),
		items: make(map[int]*list.Element),
	}
	return c
}

// Add the current element and returns the key and docID of the evicted
func (lru *LRUCache) Add(key int) (int, bool) {
	// check for existing element
	if ent, ok := lru.items[key]; ok {
		lru.list.MoveToFront(ent)
		return 0, false
	}

	// add new item
	ent := &entry{key: key}
	ety := lru.list.PushFront(ent)
	lru.items[key] = ety

	evict := lru.list.Len() > lru.size
	var evictKey int
	if evict {
		evictKey = lru.removeOldest()
	}
	return evictKey, evict
}

func (lru *LRUCache) removeOldest() int {
	ent := lru.list.Back()
	var key int
	//fmt.Println(ent, "back")
	if ent != nil {
		lru.list.Remove(ent)
		kv := ent.Value.(*entry)
		//fmt.Println(kv.key)
		key = kv.key
		delete(lru.items, kv.key)
	}
	return key
}
