package lru

import "container/list"

type Cache struct {
	mBytes int64
	nBytes int64
	queue  *list.List
	cache  map[string]*list.Element
	// callback function executed when an entry is purged
	onEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func New(mBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		mBytes:    mBytes,
		queue:     list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// Get looks up a key's value from the cache
func (c *Cache) Get(key string) (value Value, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, ok := c.cache[key]; ok {
		value = ele.Value.(*entry).value
		c.queue.MoveToFront(ele)
		return value, true
	}
	return
}

// Remove removes an item from the cache
func (c *Cache) Remove(key string) {
	if c.cache == nil {
		return
	}
	if ele, ok := c.cache[key]; ok {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(ele *list.Element) {
	e := ele.Value.(*entry)
	delete(c.cache, e.key)
	c.queue.Remove(ele)
	c.nBytes -= int64(len(e.key)) + int64(e.value.Len())
	if c.onEvicted != nil {
		c.onEvicted(e.key, e.value)
	}
}

// RemoveOldest removes the oldest item from the cache
func (c *Cache) RemoveOldest() {
	ele := c.queue.Back()
	if ele != nil {
		e := ele.Value.(*entry)
		delete(c.cache, e.key)
		c.queue.Remove(ele)
		c.nBytes -= int64(len(e.key)) + int64(e.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(e.key, e.value)
		}
	}
}

// Add adds an item to the cache (or update the value by the given key)
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.queue.MoveToFront(ele)
		e := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(e.value.Len())
		e.value = value
	} else {
		ele := c.queue.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.nBytes > c.mBytes {
		c.RemoveOldest()
	}
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	return c.queue.Len()
}
