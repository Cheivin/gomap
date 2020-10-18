package gomap

import (
	"testing"
)

func TestLinkedMap(t *testing.T) {
	m := NewLinkedMap()
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)
	m.Store("d", 4)

	t.Log(m.Load("a"))
	t.Log(m.Load("e"))
	t.Log("=======")

	t.Log(m.LoadOrStore("e", 5))
	t.Log(m.LoadOrStore("e", 5))
	t.Log("=======")

	compare := func(stored interface{}, input interface{}) interface{} {
		if stored.(int) < input.(int) {
			return input
		}
		return stored
	}
	m.StoreOrCompare("a", 3, compare)
	m.StoreOrCompare("e", 10, compare)
	m.StoreOrCompare("f", 6, compare)

	t.Log(m.Load("a"))
	t.Log(m.Load("e"))
	t.Log(m.Load("f"))
	t.Log("=======")
	m.Range(func(key interface{}, value interface{}) bool {
		t.Log("range", key, value)
		return true
	})

}
