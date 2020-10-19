package gomap

import (
	"strconv"
	"testing"
)

func TestLinkedMap_Store(t *testing.T) {
	m := NewLinkedMap()
	m.Store("1", 1)
}

func TestLinkedMap_Load(t *testing.T) {
	m := NewLinkedMap()
	m.Store("1", 1)
	t.Log(m.Load("1"))
	t.Log(m.Load("2"))
}

func TestLinkedMap_LoadOrStore(t *testing.T) {
	m := NewLinkedMap()
	t.Log(m.LoadOrStore("1", 3))
	t.Log(m.LoadOrStore("1", 3))
	t.Log(m.LoadOrStore("2", 4))
}

func TestLinkedMap_StoreOrCompare(t *testing.T) {
	m := NewLinkedMap()
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

func TestLinkedMap_Delete(t *testing.T) {
	m := NewLinkedMap()
	m.Store("1", 3)
	t.Log(m.Load("1"))
	t.Log(m.Delete("1"))
	t.Log(m.Delete("2"))
	t.Log(m.Load("1"))
}

func TestLinkedMap_Clear(t *testing.T) {
	m := NewLinkedMap()
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	t.Log(m.Clear())
	t.Log(m.Load("1"))
}

func TestLinkedMap_Range(t *testing.T) {
	m := NewLinkedMap()
	for i := 0; i < 10; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	m.Range(func(key interface{}, value interface{}) bool {
		t.Log(key, value)
		return true
	})
}

func TestLinkedMap_Destroy(t *testing.T) {
	m := NewLinkedMap()
	m.Destroy()
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	m.Load("1")
}
