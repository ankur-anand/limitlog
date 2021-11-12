package limitlog

import (
	"testing"
)

func TestLRUCache_Add(t *testing.T) {
	l := NewLRU(128)
	evictCounter := 0
	for i := 0; i < 256; i++ {
		old, evicted := l.Add(i)
		if evicted {
			evictCounter++
			if old > 127 {
				t.Errorf("wrong order of key evicted")
			}
		}
	}

	if l.list.Len() != 128 {
		t.Errorf("bad length %d", l.list.Len())
	}

	if evictCounter != 128 {
		t.Errorf("bad evict count: %v", evictCounter)
	}
}
