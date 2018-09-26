package lru

import (
	"container/list"
	"github.com/pkg/errors"
)

// LRU takes advantage of list's sequence and map's efficient locate
type LRU struct {
	ll *list.List // list.Element.Value type is of interface{}
	m  map[interface{}]*list.Element
}
type Pair struct {
	Key   interface{}
	Value interface{}
}

func (cl *LRU) Keys() []interface{} {
	keys := []interface{}{}
	for key, _ := range cl.m {
		keys = append(keys, key)
	}
	return keys
}
func (cl *LRU) Values() []interface{} {
	values := []interface{}{}
	for _, value := range cl.m {
		values = append(values, value)
	}
	return values
}

func (cl *LRU) Pairs() []Pair {
	pairs := []Pair{}
	for key, value := range cl.m {
		pairs = append(pairs, Pair{
			Key:   key,
			Value: value,
		})
	}
	return pairs
}
func (cl *LRU) AddPair(pair Pair) error {
	return cl.Add(pair.Key, pair.Value)
}

// add adds Key to the head of the linked list.
func (cl *LRU) Add(key interface{}, value interface{}) error {
	if cl.ll == nil {
		cl.ll = list.New()
		cl.m = make(map[interface{}]*list.Element)
	}
	ele := cl.ll.PushFront(Pair{
		Key:   key,
		Value: value,
	})
	if _, ok := cl.m[key]; ok {
		return errors.New("Key was already in LRU")
	}
	cl.m[key] = ele
	return nil
}

func (cl *LRU) AddOrUpdate(key interface{}, value interface{}) {
	cl.Remove(key)
	cl.Add(key, value)
}

func (cl *LRU) RemoveOldest() interface{} {
	if cl.ll == nil {
		return nil
	}
	ele := cl.ll.Back()
	pair := ele.Value.(Pair)
	cl.ll.Remove(ele)
	delete(cl.m, pair.Key)
	return pair.Value
}

// Remove removes Key from cl.
func (cl *LRU) Remove(key interface{}) {
	if ele, ok := cl.m[key]; ok {
		cl.ll.Remove(ele)
		delete(cl.m, key)
	}
}

func (cl *LRU) Find(key interface{}) (interface{}, bool) {
	e, ok := cl.m[key]
	if !ok {
		return nil, ok
	}
	return e.Value.(Pair).Value, true
}

func (cl *LRU) Peek(key interface{}) (interface{}, bool) {
	e, ok := cl.m[key]
	if ok {
		cl.Remove(key)
	}
	return e.Value.(Pair).Value, ok
}

// Len returns the number of items in the cache.
func (cl *LRU) Len() int {
	return len(cl.m)
}
