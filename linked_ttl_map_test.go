package gomap

import (
	"strconv"
	"testing"
	"time"
)

func TestLinkedTTLMap_Store(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	m.Store("1", 1)
}

func TestLinkedTTLMap_Load(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	m.Store("1", 1)
	t.Log(m.Load("1"))
	t.Log(m.Load("2"))
}

func TestLinkedTTLMap_Expiration(t *testing.T) {
	m := NewLinkedTTLMap(3*time.Second, 500*time.Millisecond, false)
	m.Store("1", 1)
	time.Sleep(3 * time.Second)
	t.Log(m.Load("1"))
}

func TestLinkedTTLMap_Expiration2(t *testing.T) {
	m := NewLinkedTTLMap(3*time.Second, 500*time.Millisecond, false)
	for i := 0; i < 10; i = i + 2 {
		m.Store(strconv.Itoa(i), i)
	}
	time.Sleep(2 * time.Second)
	for i := 1; i < 10; i = i + 2 {
		m.Store(strconv.Itoa(i), i)
	}
	time.Sleep(1 * time.Second)
	m.Range(func(key interface{}, value interface{}) bool {
		t.Log(key, value)
		return true
	})
	t.Log(m.Clear())
}

func TestLinkedTTLMap_RenewOnLoad_Load(t *testing.T) {
	m := NewLinkedTTLMap(3*time.Second, 500*time.Millisecond, true)
	m.Store("1", 1)
	t.Log(m.Load("1"))
	time.Sleep(2 * time.Second)
	t.Log(m.Load("1"))
	time.Sleep(2 * time.Second)
	t.Log(m.Load("1"))
	time.Sleep(5 * time.Second)
	t.Log(m.Load("1"))
}

func TestLinkedTTLMap_LoadOrStore(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	t.Log(m.LoadOrStore("1", 3))
	t.Log(m.LoadOrStore("1", 3))
	t.Log(m.LoadOrStore("2", 4))
}

func TestLinkedTTLMap_StoreOrCompare(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	t.Log(m.Load("1"))
	m.StoreOrCompare("1", 6, func(stored interface{}, input interface{}) interface{} {
		if stored.(int) < input.(int) {
			return input
		} else {
			return stored
		}
	})
	t.Log(m.Load("1"))

	m.Range(func(key interface{}, value interface{}) bool {
		t.Log(key, value)
		return true
	})
}

func TestLinkedTTLMap_Delete(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	t.Log(m.Load("1"))
	t.Log(m.Delete("1"))
	t.Log(m.Load("1"))
	t.Log(m.Delete("x"))
}

func TestLinkedTTLMap_Clear(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	t.Log(m.Clear())
	t.Log(m.Load("1"))
}

func TestLinkedTTLMap_Range(t *testing.T) {
	m := NewLinkedTTLMap(-1, -1, false)
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	m.Range(func(key interface{}, value interface{}) bool {
		t.Log(key, value)
		return true
	})
}

func TestLinkedTTLMap_Destroy(t *testing.T) {
	m := NewLinkedTTLMap(time.Second, time.Second, false)
	m.Destroy()
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	m.Load("1")
}
