package gomap

import (
	"errors"
	"sync"
)

type (
	LinkedMap struct {
		entryMap map[string]*linkedEntry // 缓存数据
		mu       sync.RWMutex            // 锁
		head     *linkedEntry            // 头节点
		tail     *linkedEntry            // 尾节点
	}

	linkedEntry struct {
		Entry               // 对象
		before *linkedEntry // 前一节点
		after  *linkedEntry // 后一节点
	}
)

func NewLinkedMap() *LinkedMap {
	c := &LinkedMap{
		entryMap: map[string]*linkedEntry{},
		head:     nil,
		mu:       sync.RWMutex{},
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
	entry, ok := m.entryMap[key]
	if ok {
		entry := &linkedEntry{
			Entry: Entry{
				Key:   key,
				Value: value,
			},
			before: entry.before,
			after:  entry.after,
		}
		if entry.before != nil {
			entry.before.after = entry
		} else {
			m.head = entry
		}
		if entry.after != nil {
			entry.after.before = entry
		} else {
			m.tail = entry
		}
	} else {
		entry = &linkedEntry{
			Entry: Entry{
				Key:   key,
				Value: value,
			},
			before: m.tail,
			after:  nil,
		}
		if entry.before == nil {
			m.head = entry
		} else {
			m.tail.after = entry
		}
		m.tail = entry
	}
	m.entryMap[key] = entry
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
	if item, ok := m.entryMap[key]; ok {
		delete(m.entryMap, item.Key)
		if item.after != nil {
			item.after.before = item.before
			item.after = nil
		} else {
			m.tail = item.before
		}
		if item.before != nil {
			item.before.after = item.after
			item.before = nil
		} else {
			m.head = item.after
		}
	}
	return nil
}

func (m *LinkedMap) Clear() []Entry {
	m.mu.Lock()
	if m.entryMap == nil {
		m.mu.Unlock()
		panic(errors.New(ErrMapDestroyed))
	}
	node := m.head
	m.entryMap = map[string]*linkedEntry{}
	m.head = nil
	m.tail = nil
	m.mu.Unlock()
	var entries []Entry
	for node != nil {
		entries = append(entries, node.Entry)
		if node.before != nil {
			node.before.after = nil
			node.before = nil
		}
		node = node.after
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
	m.Clear()
	m.entryMap = nil
}

func (m *LinkedMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	return len(m.entryMap)
}
