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
		object     interface{}
		expiration int64
	}
)

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
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UnixNano()

	deleted := map[string]interface{}{}
	for key, v := range m.entryMap {
		if v.expiration > 0 && now > v.expiration {
			delete(m.entryMap, key)
			deleted[key] = v.object
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
		object:     value,
		expiration: expiration,
	}
}

func (m *TTLMap) Store(key string, value interface{}) {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	m.store(key, value)
	m.mu.Unlock()
}

func (m *TTLMap) Load(key string) (value interface{}, ok bool) {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.entryMap[key]
	if ok {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			return item.object, true
		} else {
			delete(m.entryMap, key)
		}
	}
	return nil, false
}

func (m *TTLMap) LoadOrStore(key string, value interface{}) (actual interface{}, loaded bool) {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, ok := m.entryMap[key]; ok {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			return item.object, true
		}
	}
	m.store(key, value)
	return value, false
}

func (m *TTLMap) StoreIfPresent(key string, value interface{}, compare func(stored interface{}, input interface{}) interface{}) {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, ok := m.entryMap[key]; ok {
		if !item.expired() {
			item.renew(m.expiration)
			if compare != nil {
				item.object = compare(item.object, value)
			}
			return
		}
	}
	// 存入值
	m.store(key, value)
}

func (m *TTLMap) Delete(key string) interface{} {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, ok := m.entryMap[key]; ok {
		delete(m.entryMap, key)
		if !val.expired() {
			return val.object
		}
	}
	return nil
}

func (m *TTLMap) Clear() map[string]interface{} {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	now := time.Now().UnixNano()
	if len(m.entryMap) == 0 {
		return nil
	}
	deleted := make(map[string]interface{}, len(m.entryMap))
	m.entryMap = map[string]*ttlEntry{}
	m.mu.Unlock()
	for key, v := range m.entryMap {
		if v.expiration <= 0 || now <= v.expiration {
			deleted[key] = v.object
		}
	}
	return deleted
}

func (m *TTLMap) Range(f func(key interface{}, value interface{}) bool) {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for key, item := range m.entryMap {
		if !item.expired() {
			if m.renewOnLoad {
				item.renew(m.expiration)
			}
			if !f(key, item.object) {
				break
			}
		}
	}
}

func (m *TTLMap) Destroy() {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exit <- true
	m.entryMap = nil
}

func (m *TTLMap) Size() int {
	if m.entryMap == nil {
		panic(errors.New(MapDestroyed))
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entryMap)
}
