package _map

import (
	"fmt"
	"sync"
)

type Map struct {
	m    map[string]any
	lock sync.Mutex
}

func NewMap() *Map {
	return &Map{
		m: make(map[string]any),
	}
}

func (m *Map) Put(key string, value any) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, exist := m.m[key]; exist {
		return fmt.Errorf("the key %s has existed", key)
	}
	m.m[key] = value
	return nil
}

func (m *Map) Del(key string) {
	m.lock.Lock()
	delete(m.m, key)
	m.lock.Unlock()
}

func (m *Map) Get(key string) any {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, exist := m.m[key]; !exist {
		return nil
	}
	return m.m[key]
}

func (m *Map) Update(key string, val any) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[key] = val
}

func (m *Map) UpdateMap(newMap *map[string]any) {
	m.lock.Lock()
	m.m = make(map[string]any)
	for key, val := range *newMap {
		m.m[key] = val
	}
	m.lock.Unlock()
}

func (m *Map) Size() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.m)
}

func (m *Map) Copy() map[string]any {
	m.lock.Lock()
	defer m.lock.Unlock()
	result := make(map[string]any)
	for key, val := range m.m {
		result[key] = val
	}
	return result
}

func (m *Map) CheckIfAllExist(target *map[string]any) {
	m.lock.Lock()
	for key, _ := range *target {
		if _, exist := m.m[key]; !exist {
			delete(*target, key)
		}
	}
	m.lock.Unlock()
}
