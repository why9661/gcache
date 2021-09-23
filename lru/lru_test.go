package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest failed.")
	}
}

func TestOnEvicted(t *testing.T) {
	var keys []string
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(10), callback)
	lru.Add("k1", String("value1"))
	lru.Add("k2", String("value2"))
	lru.Add("k3", String("value3"))

	expected := []string{"k1", "k2"}

	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("OnEvicted function failed.")
	}
}
