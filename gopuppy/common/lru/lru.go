package lru

import (
	"container/list"
	"errors"
)

type EvictCallback func(key interface{}, value interface{})

type LruCache struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	onEvict   EvictCallback
}

type entry struct {
	key   interface{}
	value interface{}
}

func NewLruCache(size int, onEvict EvictCallback) (*LruCache, error) {
	if size <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	c := &LruCache{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

func (c *LruCache) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

func (c *LruCache) Add(key, value interface{}) (evicted bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		return false
	}

	ent := &entry{key, value}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry

	evict := c.evictList.Len() > c.size
	if evict {
		c.removeOldest()
	}
	return evict
}

func (c *LruCache) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return
}

func (c *LruCache) Contains(key interface{}) (ok bool) {
	_, ok = c.items[key]
	return ok
}

func (c *LruCache) Peek(key interface{}) (value interface{}, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

func (c *LruCache) Remove(key interface{}) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent, true)
		return true
	}
	return false
}

func (c *LruCache) RemoveWithoutCallback(key interface{}) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent, false)
		return true
	}
	return false
}

func (c *LruCache) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent, true)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

func (c *LruCache) GetOldest() (key interface{}, value interface{}, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

func (c *LruCache) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

func (c *LruCache) Len() int {
	return c.evictList.Len()
}

func (c *LruCache) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent, true)
	}
}

func (c *LruCache) removeElement(e *list.Element, needEvictCallback bool) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	if needEvictCallback && c.onEvict != nil {
		c.onEvict(kv.key, kv.value)
	}
}

func (c *LruCache) ForEach(callback func(value interface{})) {
	for _, ent := range c.items {
		callback(ent.Value.(*entry).value)
	}
}
