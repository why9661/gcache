package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunc func([]byte) uint32

type CHash struct {
	hash     HashFunc
	replicas int
	keys     []int
	hashMap  map[int]string
}

func New(replicas int, fn HashFunc) *CHash {
	m := &CHash{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *CHash) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *CHash) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[index%len(m.keys)]]
}
