package gomap

import (
	"strconv"
	"testing"
	"time"
)

func TestTTLMap_Store(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	m.Store("1", 1)
}

func TestTTLMap_Load(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	m.Store("1", 1)
	t.Log(m.Load("1"))
	t.Log(m.Load("2"))
}

func TestTTLMap_Expiration(t *testing.T) {
	m := NewTTLMap(3*time.Second, 500*time.Millisecond, false)
	m.Store("1", 1)
	time.Sleep(3 * time.Second)
	t.Log(m.Load("1"))
}

func TestTTLMap_RenewOnLoad_Load(t *testing.T) {
	m := NewTTLMap(3*time.Second, 500*time.Millisecond, true)
	m.Store("1", 1)
	t.Log(m.Load("1"))
	time.Sleep(2 * time.Second)
	t.Log(m.Load("1"))
	time.Sleep(2 * time.Second)
	t.Log(m.Load("1"))
	time.Sleep(5 * time.Second)
	t.Log(m.Load("1"))
}

func TestTTLMap_LoadOrStore(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	t.Log(m.LoadOrStore("1", 3))
	t.Log(m.LoadOrStore("1", 3))
	t.Log(m.LoadOrStore("2", 4))
}

func TestTTLMap_StoreOrCompare(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	m.StoreOrCompare("1", 3, nil)
	t.Log(m.Load("1"))
	m.StoreOrCompare("1", 6, func(stored interface{}, input interface{}) interface{} {
		if stored.(int) < input.(int) {
			return input
		} else {
			return stored
		}
	})
	t.Log(m.Load("1"))
}

func TestTTLMap_Delete(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	m.Store("1", 3)
	t.Log(m.Load("1"))
	t.Log(m.Delete("1"))
	t.Log(m.Delete("2"))
	t.Log(m.Load("1"))
}

func TestTTLMap_Clear(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	t.Log(m.Clear())
	t.Log(m.Load("1"))
}

func TestTTLMap_Range(t *testing.T) {
	m := NewTTLMap(-1, -1, false)
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	m.Range(func(key interface{}, value interface{}) bool {
		t.Log(key, value)
		return true
	})
}

func TestTTLMap_Destroy(t *testing.T) {
	m := NewTTLMap(time.Second, time.Second, false)
	m.Destroy()
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	m.Load("1")
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
