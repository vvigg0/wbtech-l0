package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type TTLMap struct {
	m map[string]*CacheItem
	l sync.RWMutex
}

func NewTTLMap() *TTLMap {
	return &TTLMap{
		m: make(map[string]*CacheItem),
	}
}

func (t *TTLMap) Set(key string, value interface{}) {
	t.l.Lock()
	defer t.l.Unlock()
	expiration := time.Now().Add(2 * time.Minute).Unix()
	t.m[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

func (t *TTLMap) Get(key string) (interface{}, bool) {
	t.l.RLock()
	defer t.l.RUnlock()
	value, ok := t.m[key]
	if !ok {
		return nil, false
	}
	if time.Now().Unix() > value.Expiration {
		t.Delete(key)
		return nil, false
	}
	return value.Value, true
}

func (t *TTLMap) Delete(key string) {
	t.l.Lock()
	defer t.l.Unlock()
	delete(t.m, key)
}

func (t *TTLMap) Clean(interval time.Duration) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		t.l.Lock()
		for key, value := range t.m {
			if time.Now().Unix() > value.Expiration {
				delete(t.m, key)
			}
		}
		t.l.Unlock()
	}
}
