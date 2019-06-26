package main

import (
	"fmt"
	"sync"
	"time"
)

type SyncMap_string_bool struct {
	mu	sync.Mutex
	m	map[string]bool
}

func (sm *SyncMap_string_bool) Delete(key string)	{ sm.mu.Lock(); delete(sm.m, key); sm.mu.Unlock() }

func (sm *SyncMap_string_bool) Load(key string) (value bool, ok bool) {
	sm.mu.Lock()
	value, ok = sm.m[key]
	sm.mu.Unlock()
	return
}

func (sm *SyncMap_string_bool) LoadOrStore(key string, value bool) (actual bool, loaded bool) {
	sm.mu.Lock()
	actual, loaded = sm.m[key]
	if !loaded {
		sm.m[key] = value
		actual = value
	}

	sm.mu.Unlock()
	return
}

func (sm *SyncMap_string_bool) Range(f func(key string, value bool) bool) {
	sm.mu.Lock()
	for k, v := range sm.m {
		if !f(k, v) {
			break
		}
	}

	sm.mu.Unlock()
}

func (sm *SyncMap_string_bool) Store(key string, value bool) {
	sm.mu.Lock()
	sm.m[key] = value
	sm.mu.Unlock()
}
func MakeSyncMap_string_bool() *SyncMap_string_bool {
	return &SyncMap_string_bool{m: make(map[string]bool)}
}

func MarkAll_string(done chan<- bool, sm *SyncMap_string_bool, values ...string) {
	for _, val := range values {
		time.Sleep(time.Second / 10)
		sm.Store(val, true)
	}

	done <- true
}
func main() {
	marked := MakeSyncMap_string_bool()
	done := make(chan bool)
	go MarkAll_string(done, marked, "A", "B", "C", "D")
	go MarkAll_string(done, marked, "E", "F", "G", "H")
	go MarkAll_string(done, marked, "I", "J", "K", "L")
	for i := 0; i < 3; i++ {
		<-done
	}

	marked.Range(func(key string, value bool) bool { fmt.Println(key); return true })
}
