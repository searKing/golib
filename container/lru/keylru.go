package lru

import "container/list"

// LRU takes advantage of list's sequence and map's efficient locate
type KeyLRU struct {
	ll *list.List // list.Element.Value type is of interface{}
	m  map[interface{}]*list.Element
}

func (cl *KeyLRU) Keys() []interface{} {
	keys := []interface{}{}
	for key, _ := range cl.m {
		keys = append(keys, key)
	}
	return keys
}

// add adds Key to the head of the linked list.
func (cl *KeyLRU) Add(key interface{}) {
	if cl.ll == nil {
		cl.ll = list.New()
		cl.m = make(map[interface{}]*list.Element)
	}
	ele := cl.ll.PushFront(key)
	if _, ok := cl.m[key]; ok {
		panic("persistConn was already in LRU")
	}
	cl.m[key] = ele
}
func (cl *KeyLRU) AddOrUpdate(key interface{}, value interface{}) {
	cl.Remove(key)
	cl.Add(key)
}

func (cl *KeyLRU) RemoveOldest() interface{} {
	if cl.ll == nil {
		return nil
	}
	ele := cl.ll.Back()
	key := ele.Value.(interface{})
	cl.ll.Remove(ele)
	delete(cl.m, key)
	return key
}

// Remove removes Key from cl.
func (cl *KeyLRU) Remove(key interface{}) {
	if ele, ok := cl.m[key]; ok {
		cl.ll.Remove(ele)
		delete(cl.m, key)
	}
}

func (cl *KeyLRU) Find(key interface{}) (interface{}, bool) {
	e, ok := cl.m[key]
	return e, ok
}

func (cl *KeyLRU) Peek(key interface{}) (interface{}, bool) {
	e, ok := cl.m[key]
	if ok {
		cl.Remove(key)
	}
	return e, ok
}

// Len returns the number of items in the cache.
func (cl *KeyLRU) Len() int {
	return len(cl.m)
}
