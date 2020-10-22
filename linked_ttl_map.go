package gomap

import (
	"errors"
	"sync"
	"time"
)

type (
	LinkedTTLMap struct {
		entryMap    map[string]*linkedTTLEntry // 缓存数据
		mu          *sync.RWMutex              // 锁
		exit        chan bool                  // 退出标志
		gcInterval  time.Duration              // 清理周期
		expiration  time.Duration              // 过期时间
		renewOnLoad bool                       // 读取时续租时间
		head        *linkedTTLEntry            // 头节点
		tail        *linkedTTLEntry
	}

	linkedTTLEntry struct {
		*ttlEntry
		before *linkedTTLEntry // 前一节点
		after  *linkedTTLEntry // 后一节点
	}
)

func NewLinkedTTLMap(expiration, gcInterval time.Duration, renewOnLoad bool) *LinkedTTLMap {
	m := &LinkedTTLMap{
		expiration:  expiration,
		gcInterval:  gcInterval,
		entryMap:    map[string]*linkedTTLEntry{},
		mu:          &sync.RWMutex{},
		exit:        make(chan bool),
		renewOnLoad: renewOnLoad,
	}
	if expiration > 0 {
		go m.gcLoop()
	}
	return m
}

//gcLoop 过期清理轮询
func (m *LinkedTTLMap) gcLoop() {
	if m.gcInterval <= 0 {
		m.gcInterval = 100 * time.Millisecond
	}
	if m.expiration <= 0 {
		return
	}
	ticker := time.NewTicker(m.gcInterval)
	for {
		select {
		case <-ticker.C:
			m.DeleteExpired()
		case <-m.exit:
			ticker.Stop()
			return
		}
	}
}

//DeleteExpired 删除过期数据项
func (m *LinkedTTLMap) DeleteExpired() []Entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	var entries []Entry
	for _, v := range m.entryMap {
		if v.expired() {
			m.delete(v)
			entries = append(entries, v.Entry)
		}
	}
	return entries
}

func (m *LinkedTTLMap) store(key string, value interface{}) {
	var expiration int64
	if m.expiration > 0 {
		expiration = time.Now().Add(m.expiration).UnixNano()
	} else {
		expiration = -1
	}
	entry, ok := m.entryMap[key]
	if ok {
		entry := &linkedTTLEntry{
			ttlEntry: &ttlEntry{
				Entry: Entry{
					Key:   key,
					Value: value,
				},
				expiration: expiration,
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
		entry = &linkedTTLEntry{
			ttlEntry: &ttlEntry{
				Entry: Entry{
					Key:   key,
					Value: value,
				},
				expiration: expiration,
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

func (m *LinkedTTLMap) Store(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	m.store(key, value)
}

func (m *LinkedTTLMap) Load(key string) (value interface{}, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	item, ok := m.entryMap[key]
	if ok {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			return item.Value, true
		} else {
			m.delete(item)
		}
	}
	return nil, false
}

func (m *LinkedTTLMap) delete(item *linkedTTLEntry) interface{} {
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
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
	return item.Value
}

func (m *LinkedTTLMap) LoadOrStore(key string, value interface{}) (actual interface{}, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	if item, ok := m.entryMap[key]; ok {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			return item.Value, true
		}
	}
	m.store(key, value)
	return value, false

}

func (m *LinkedTTLMap) StoreOrCompare(key string, value interface{}, compare func(stored interface{}, input interface{}) interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}

	if item, ok := m.entryMap[key]; ok {
		if !item.expired() {
			item.renew(m.expiration)
			if compare != nil {
				item.Value = compare(item.Value, value)
			}
			m.entryMap[key] = item
			return
		}
	}
	// 存入值
	m.store(key, value)
}

func (m *LinkedTTLMap) Delete(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	if item, ok := m.entryMap[key]; ok {
		return m.delete(item)
	}
	return nil
}

func (m *LinkedTTLMap) Clear() []Entry {
	m.mu.Lock()
	if m.entryMap == nil {
		m.mu.Unlock()
		panic(errors.New(ErrMapDestroyed))
	}
	node := m.head
	m.entryMap = map[string]*linkedTTLEntry{}
	m.head = nil
	m.tail = nil
	m.mu.Unlock()
	var entries []Entry
	for node != nil {
		if !node.expired() {
			entries = append(entries, node.Entry)
		}
		if node.before != nil {
			node.before.after = nil
			node.before = nil
		}
		node = node.after
	}
	return entries
}

func (m *LinkedTTLMap) Range(f func(key interface{}, value interface{}) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	node := m.head
	for node != nil {
		if !node.expired() {
			if !f(node.Key, node.Value) {
				break
			}
		}
		node = node.after
	}
}

func (m *LinkedTTLMap) Destroy() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	m.Clear()
	m.entryMap = nil
	close(m.exit)
}

func (m *LinkedTTLMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(ErrMapDestroyed))
	}
	return len(m.entryMap)
}
