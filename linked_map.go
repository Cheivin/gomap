package gomap

import (
	"errors"
	"sync"
)

type (
	LinkedMap struct {
		entryMap map[string]*linkedEntry // 缓存数据
		mu       *sync.RWMutex           // 锁
		head     *linkedEntry            // 头节点
		tail     *linkedEntry            // 尾节点
	}

	linkedEntry struct {
		*Entry              // 对象
		before *linkedEntry // 前一节点
		after  *linkedEntry // 后一节点
	}
)

func NewLinkedMap() *LinkedMap {
	c := &LinkedMap{
		entryMap: map[string]*linkedEntry{},
		head:     nil,
		mu:       &sync.RWMutex{},
	}
	return c
}

func (m *LinkedMap) Store(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	m.store(key, value)
}

func (m *LinkedMap) store(key string, value interface{}) {
	entry := &linkedEntry{
		before: m.tail,
		after:  nil,
		Entry: &Entry{
			Key:   key,
			Value: value,
		},
	}
	m.entryMap[key] = entry

	if m.tail == nil {
		m.head = entry
	} else {
		m.tail.after = entry
	}
	m.tail = entry
}

func (m *LinkedMap) Load(key string) (value interface{}, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	item, ok := m.entryMap[key]
	if ok {
		return item.Value, true
	}
	return nil, false
}

func (m *LinkedMap) LoadOrStore(key string, value interface{}) (actual interface{}, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	if item, ok := m.entryMap[key]; ok {
		return item.Value, true
	}
	m.store(key, value)
	return value, false
}

func (m *LinkedMap) StoreOrCompare(key string, value interface{}, compare func(stored interface{}, input interface{}) interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}

	if item, ok := m.entryMap[key]; ok {
		if compare != nil {
			item.Value = compare(item.Value, value)
		}
		m.entryMap[key] = item
		return
	}
	// 存入值
	m.store(key, value)
}

func (m *LinkedMap) Delete(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	if val, ok := m.entryMap[key]; ok {
		delete(m.entryMap, key)
		return val.Value
	}
	return nil
}

func (m *LinkedMap) Clear() []Entry {
	m.mu.Lock()
	if m.entryMap == nil {
		m.mu.Unlock()
		panic(errors.New(ErrMapDestroyed))
	}
	deleted := m.entryMap
	m.entryMap = map[string]*linkedEntry{}
	m.mu.Unlock()
	var entries []Entry
	for _, v := range deleted {
		entries = append(entries, *v.Entry)
	}
	return entries
}

func (m *LinkedMap) Range(f func(key interface{}, value interface{}) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	node := m.head
	for node != nil {
		if !f(node.Key, node.Value) {
			break
		}
		node = node.after
	}
}

func (m *LinkedMap) Destroy() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	m.entryMap = nil
	m.head = nil
	m.tail = nil
}

func (m *LinkedMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	return len(m.entryMap)
}
