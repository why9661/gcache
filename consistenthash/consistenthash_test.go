package consistenthash

import (
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	chash := New(3, func(b []byte) uint32 {
		i, _ := strconv.Atoi(string(b))
		return uint32(i)
	})

	chash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"5":  "6",
		"13": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if chash.Get(k) != v {
			t.Fatalf("result should be %s-%s", k, v)
		}
	}

	chash.Add("8")
	testCases["27"] = "8"

	for k, v := range testCases {
		if chash.Get(k) != v {
			t.Fatalf("result should be %s-%s", k, v)
		}
	}
}
