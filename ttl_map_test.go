package gomap

import (
	"strconv"
	"testing"
)

func TestTTLMap_Store(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	m.store("1", 1)
}

func BenchmarkTTLMap_Store(b *testing.B) {
	m := NewTTLMap(-1, -1, false)
	var keys []string
	for i := 0; i < b.N; i++ {
		keys = append(keys, strconv.Itoa(i))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i, key := range keys {
		m.Store(key, i)
	}
}
