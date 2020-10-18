package gomap

import (
	"errors"
	"sync"
	"time"
)

type (
	TTLMap struct {
		entryMap    map[string]*ttlEntry // 缓存数据
		mu          *sync.RWMutex        // 锁
		exit        chan bool            // 退出标志
		gcInterval  time.Duration        // 清理周期
		expiration  time.Duration        // 过期时间
		renewOnLoad bool                 // 读取时续租时间
	}

	ttlEntry struct {
		*Entry
		expiration int64
	}
)

func NewTTLMap(expiration, gcInterval time.Duration, renewOnLoad bool) *TTLMap {
	m := &TTLMap{
		expiration:  expiration,
		gcInterval:  gcInterval,
		entryMap:    map[string]*ttlEntry{},
		mu:          &sync.RWMutex{},
		exit:        make(chan bool),
		renewOnLoad: renewOnLoad,
	}
	if expiration > 0 {
		go m.gcLoop()
	}
	return m
}

func (e *ttlEntry) expired() bool {
	if e.expiration <= 0 {
		return false
	}
	return time.Now().UnixNano() > e.expiration
}

func (e *ttlEntry) renew(expiration time.Duration) {
	if e.expired() {
		return
	}
	e.expiration = time.Now().Add(expiration).UnixNano()
}

//gcLoop 过期清理轮询
func (m *TTLMap) gcLoop() {
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
func (m *TTLMap) DeleteExpired() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	now := time.Now().UnixNano()
	deleted := map[string]interface{}{}
	for key, v := range m.entryMap {
		if v.expiration > 0 && now > v.expiration {
			delete(m.entryMap, key)
			deleted[key] = v.Value
		}
	}
	return deleted
}

func (m *TTLMap) store(key string, value interface{}) {
	var expiration int64
	if m.expiration > 0 {
		expiration = time.Now().Add(m.expiration).UnixNano()
	} else {
		expiration = -1
	}
	m.entryMap[key] = &ttlEntry{
		Entry: &Entry{
			Key:   key,
			Value: value,
		},
		expiration: expiration,
	}
}

func (m *TTLMap) Store(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.store(key, value)
}

func (m *TTLMap) Load(key string) (value interface{}, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	item, ok := m.entryMap[key]
	if ok {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			return item.Value, true
		} else {
			delete(m.entryMap, key)
		}
	}
	return nil, false
}

func (m *TTLMap) LoadOrStore(key string, value interface{}) (actual interface{}, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
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

func (m *TTLMap) StoreOrCompare(key string, value interface{}, compare func(stored interface{}, input interface{}) interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
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

func (m *TTLMap) Delete(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	if val, ok := m.entryMap[key]; ok {
		delete(m.entryMap, key)
		if !val.expired() {
			return val.Value
		}
	}
	return nil
}

func (m *TTLMap) Clear() []Entry {
	m.mu.Lock()
	if m.entryMap == nil {
		m.mu.Unlock()
		panic(errors.New(MapDestroyed))
	}
	now := time.Now().UnixNano()
	deleted := m.entryMap
	m.entryMap = map[string]*ttlEntry{}
	m.mu.Unlock()
	var entries []Entry
	for _, v := range deleted {
		if v.expiration <= 0 || now <= v.expiration {
			entries = append(entries, *v.Entry)
		}
	}
	return entries
}

func (m *TTLMap) Range(f func(key interface{}, value interface{}) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	for key, item := range m.entryMap {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			if !f(key, item.Value) {
				break
			}
		}
	}
}

func (m *TTLMap) Destroy() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.exit <- true
	m.entryMap = nil
}

func (m *TTLMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	return len(m.entryMap)
}
