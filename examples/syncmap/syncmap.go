package main

import (
	"fmt"
	"sync"
	"time"
)

// SyncMap is a generic hash-map usable from multiple goroutines simultaneously.
//
// The API imitates the API of sync.Map.
//
// This is a dummy implementation to demonstrate the typing capabilities. This is not
// an example of an efficient implementation of a SyncMap.
type SyncMap(type K eq, type V) struct {
	mu sync.Mutex
	m  map[K]V
}

// MakeSyncMap creates a new, empty SyncMap.
//
// I know it's better to make the zero value useful, this is just to better demonstrate
// the unnamed type parmeters syntax.
func MakeSyncMap(type K eq, type V) *SyncMap(K, V) {
	return &SyncMap(K, V){
		m: make(map[K]V),
	}
}

// Delete deletes the value for a key. 
func (sm *SyncMap(type K eq, type V)) Delete(key K) {
	sm.mu.Lock()
	delete(sm.m, key)
	sm.mu.Unlock()
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map. 
func (sm *SyncMap(type K eq, type V)) Load(key K) (value V, ok bool) {
	sm.mu.Lock()
	value, ok = sm.m[key]
	sm.mu.Unlock()
	return
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored. 
func (sm *SyncMap(type K eq, type V)) LoadOrStore(key K, value V) (actual V, loaded bool) {
	sm.mu.Lock()
	actual, loaded = sm.m[key]
	if !loaded {
		sm.m[key] = value
		actual = value
	}
	sm.mu.Unlock()
	return
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration. 
func (sm *SyncMap(type K eq, type V)) Range(f func(key K, value V) bool) {
	sm.mu.Lock()
	for k, v := range sm.m {
		if !f(k, v) {
			break
		}
	}
	sm.mu.Unlock()
}

//  Store sets the value for a key.
func (sm *SyncMap(type K eq, type V)) Store(key K, value V) {
	sm.mu.Lock()
	sm.m[key] = value
	sm.mu.Unlock()
}

func MarkAll(done chan<- bool, sm *SyncMap(type T eq, bool), values ...T) {
	for _, val := range values {
		time.Sleep(time.Second / 10)
		sm.Store(val, true)
	}
	done <- true
}

func main() {
	marked := MakeSyncMap(string, bool)
	done := make(chan bool)

	go MarkAll(done, marked, "A", "B", "C", "D")
	go MarkAll(done, marked, "E", "F", "G", "H")
	go MarkAll(done, marked, "I", "J", "K", "L")

	for i := 0; i < 3; i++ {
		<-done
	}

	marked.Range(func(key string, value bool) bool {
		fmt.Println(key)
		return true
	})
}