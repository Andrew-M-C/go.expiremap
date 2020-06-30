// Package expiremap is a simple thread-safe cache like sync.Map. Value in this map will automatically
// be deleted after expiration time.
//
// This package acts like go-cache(github.com/patrickmn/go-cache), but much faster. The limitations are:
// expiration could not changed once the map is established, and the actual expiration time for each
// value may have an error of up to one second. These are the cost of increased effenciency.
package expiremap

import (
	"container/list"
	"runtime"
	"sync"
	"time"
)

const (
	_defaultExpiration      = 5 * time.Minute
	_defaultCleanupInterval = 5 * time.Minute
)

type item struct {
	e      *list.Element
	key    interface{}
	expire time.Time
}

// Map is the main instance of this package.
type Map struct {
	m *mp
}

type mp struct {
	storage    sync.Map
	agingElems map[interface{}]*item
	agingList  *list.List
	expiration time.Duration
	newer      chan *item
	stop       chan bool
}

// New initialize and return an expiremap instance with given expiration for each element.
//
// The exact expiration time for each element may have up to one second longer because the
// map checks expiration every one second.
func New(expiration time.Duration) *Map {
	if expiration <= 0 {
		expiration = _defaultExpiration
	}

	m := &mp{
		agingElems: make(map[interface{}]*item),
		agingList:  list.New(),
		expiration: expiration,
		newer:      make(chan *item),
		stop:       make(chan bool),
	}

	ret := &Map{m: m}
	runCleanup(m)
	runtime.SetFinalizer(ret, stopCleanup)

	return ret
}

func (m *mp) run() {
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-timer.C:
			m.cleanExpires()
			timer = time.NewTimer(time.Second)
		case <-m.stop:
			// log.Printf("stop")
			return
		case itm := <-m.newer:
			m.appendNewOne(itm)
			m.cleanExpires()
		}
	}
}

func (m *mp) cleanExpires() {
	now := time.Now()
	for {
		e := m.agingList.Back()
		if nil == e {
			// nil list
			break
		}
		itm := e.Value.(*item)
		if itm.expire.After(now) {
			// no expires
			// log.Printf("itm %s not expire %v", itm.key, itm.expire)
			break
		}
		// log.Printf("clean %s (%v)", itm.key, itm.expire)
		m.storage.Delete(itm.key)
		delete(m.agingElems, itm.key)
		m.agingList.Remove(e)
	}
	return
}

func (m *mp) appendNewOne(one *item) {
	if itm, exist := m.agingElems[one.key]; exist {
		// log.Printf("extend one %s -> %v", one.key, one.expire)
		itm.expire = one.expire
		m.agingList.MoveToFront(itm.e)
		return
	}
	// trully a new one
	// log.Printf("new expiration: %v", one.expire)
	one.e = m.agingList.PushFront(one)
	m.agingElems[one.key] = one
	return
}

func runCleanup(m *mp) {
	go m.run()
}

func stopCleanup(m *Map) {
	// log.Printf("should stop")
	m.m.stop <- true
}

// Expiration returns the expiration time for each values in this map.
func (m *Map) Expiration() time.Duration {
	return m.m.expiration
}

// Store stores an value with default expiration.
func (m *Map) Store(key, value interface{}) {
	m.m.storage.Store(key, value)
	itm := item{
		key:    key,
		expire: time.Now().Add(m.m.expiration),
	}
	m.m.newer <- &itm
	return
}

// Load returns the value stored in the map for a key, or nil if no value is present. The ok result indicates whether key exists in the map.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	return m.m.storage.Load(key)
}
