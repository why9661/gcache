package lru

import "container/list"

type Lru struct {
	mBytes int64
	nBytes int64
	queue  *list.List
	cache  map[string]*list.Element
	//callback function executed when an entry is purged
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(mBytes int64, onEvicted func(string, Value)) *Lru {
	return &Lru{
		mBytes:    mBytes,
		queue:     list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get looks up a key's value from the cache
func (l *Lru) Get(key string) (value Value, ok bool) {
	if l.cache == nil {
		return
	}
	if ele, ok := l.cache[key]; ok {
		value = ele.Value.(*entry).value
		l.queue.MoveToFront(ele)
		return value, true
	}
	return
}

// Remove removes a key-value pair from the cache
func (l *Lru) Remove(key string) {
	if l.cache == nil {
		return
	}
	if ele, ok := l.cache[key]; ok {
		l.removeElement(ele)
	}
}

func (l *Lru) removeElement(ele *list.Element) {
	e := ele.Value.(*entry)
	delete(l.cache, e.key)
	l.queue.Remove(ele)
	l.nBytes -= int64(len(e.key)) + int64(e.value.Len())
	if l.OnEvicted != nil {
		l.OnEvicted(e.key, e.value)
	}
}

//RemoveOldest removes the oldest item from the cache
func (l *Lru) RemoveOldest() {
	ele := l.queue.Back()
	if ele != nil {
		e := ele.Value.(*entry)
		delete(l.cache, e.key)
		l.queue.Remove(ele)
		l.nBytes -= int64(len(e.key)) + int64(e.value.Len())
		if l.OnEvicted != nil {
			l.OnEvicted(e.key, e.value)
		}
	}
}

//Add adds a key-value pair to the cache (or update the provided key's value)
func (l *Lru) Add(key string, value Value) {
	if ele, ok := l.cache[key]; ok {
		l.queue.MoveToFront(ele)
		e := ele.Value.(*entry)
		l.nBytes += int64(value.Len()) - int64(e.value.Len())
		e.value = value
	} else {
		ele := l.queue.PushFront(&entry{key: key, value: value})
		l.cache[key] = ele
		l.nBytes += int64(len(key)) + int64(value.Len())
	}
	for l.nBytes > l.mBytes {
		l.RemoveOldest()
	}
}

//Len returns the number of items in the cache
func (l *Lru) Len() int {
	return l.queue.Len()
}
