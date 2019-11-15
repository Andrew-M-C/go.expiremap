package expiremap

import (
	"runtime"
	"sync"
	"time"
)

const (
	_defaultExpiration      = 5 * time.Minute
	_defaultCleanupInterval = 5 * time.Minute
)

type item struct {
	key    interface{}
	value  interface{}
	expire time.Time
}

// Map is the main instance of this package
type Map struct {
	m *mp
}

type mp struct {
	m             map[interface{}]*item
	mu            sync.RWMutex
	dftExp        time.Duration
	cleanInterval time.Duration
	stop          chan bool
}

// New initialize and return an expiremap instance
func New(defaultExpiration, cleanupInterval time.Duration) *Map {
	if defaultExpiration <= 0 {
		defaultExpiration = _defaultExpiration
	}
	if cleanupInterval <= 0 {
		cleanupInterval = _defaultCleanupInterval
	}

	m := &mp{
		m:             make(map[interface{}]*item),
		dftExp:        defaultExpiration,
		cleanInterval: cleanupInterval,
		stop:          make(chan bool),
	}

	ret := &Map{m: m}
	runCleanup(m)
	runtime.SetFinalizer(ret, stopCleanup)

	return ret
}

func (m *mp) run() {
	ticker := time.NewTicker(m.cleanInterval)
	for {
		select {
		case <-m.stop:
			// log.Printf("stop")
			ticker.Stop()
			return
		case <-ticker.C:
			// log.Printf("tick")
			m.deleteExpired()
		}
	}
}

func (m *mp) deleteExpired() {
	now := time.Now()

	m.mu.Lock()
	for k, itm := range m.m {
		if itm.expire.Before(now) {
			delete(m.m, k)
		}
	}
	m.mu.Unlock()

	return
}

func runCleanup(m *mp) {
	go m.run()
}

func stopCleanup(m *Map) {
	// log.Printf("should stop")
	m.m.stop <- true
}

// DefaultExpiration returns the default expiration.
func (m *Map) DefaultExpiration() time.Duration {
	return m.m.dftExp
}

// Store stores an value with default expiration.
func (m *Map) Store(key, value interface{}) {
	m.StoreWithExpiration(key, value, m.m.dftExp)
}

// StoreWithExpiration stores an value with given expiration.
func (m *Map) StoreWithExpiration(key, value interface{}, expiration time.Duration) {
	expire := time.Now().Add(expiration)

	m.m.mu.Lock()
	m.m.m[key] = &item{
		key:    key,
		value:  value,
		expire: expire,
	}
	m.m.mu.Unlock()
}

// Load returns the value stored in the map for a key, or nil if no value is present. The ok result indicates whether value was found in the map.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	m.m.mu.RLock()

	itm, ok := m.m.m[key]
	if false == ok {
		// do nothing
	} else if itm.expire.After(time.Now()) {
		value = itm.value
	} else {
		ok = false
	}

	m.m.mu.RUnlock()
	return
}

// Delete deletes the value for a key.
func (m *Map) Delete(key interface{}) {
	m.m.mu.Lock()
	delete(m.m.m, key)
	m.m.mu.Unlock()
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores with default expiration and returns the given value. The loaded result is true if the value was loaded, false if stored.
// func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
// 	return m.LoadOrStoreWithExpiration(key, value, m.DefaultExpiration())
// }

// LoadOrStoreWithExpiration returns the existing value for the key if present. Otherwise, it stores with given expiration and returns the given value. The loaded result is true if the value was loaded, false if stored.
// func (m *Map) LoadOrStoreWithExpiration(key, value interface{}, expiration time.Duration) (actual interface{}, loaded bool) {
// 	newItm := &item{
// 		key:    key,
// 		value:  value,
// 		expire: time.Now().Add(expiration),
// 	}
// 	intf, loaded := m.m.syncMap.LoadOrStore(key, newItm)
// 	if false == loaded {
// 		return value, false
// 	}
// 	if itm := intf.(*item); itm.expire.After(time.Now()) {
// 		return itm.value, true
// 	}

// 	m.m.syncMap.Store(key, newItm)
// 	return value, false
// }

// Range calls f sequentially for each key and value present in the map. If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's contents: no key will be visited more than once, but if the value for any key is stored or deleted concurrently, Range may reflect any mapping for that key from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns false after a constant number of calls.
// func (m *Map) Range(f func(key, value interface{}) bool) {
// 	m.m.syncMap.Range(f)
// }
