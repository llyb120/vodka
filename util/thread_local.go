package util

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ThreadLocal struct {
	storage          sync.Map
	activeGoroutines sync.Map
	cleanup          *time.Ticker
}

type valueWrapper struct {
	value interface{}
}

var tl *ThreadLocal

// var once sync.Once

func NewThreadLocal(autoClean bool) *ThreadLocal {
	// 做一个单例，要线程安全
	// once.Do(func() {
	tl = &ThreadLocal{
		cleanup: time.NewTicker(time.Minute * 5),
	}
	if autoClean {
		go tl.cleanupRoutine()
	}
	// })
	return tl
}

func (t *ThreadLocal) Set(value interface{}) {
	id := getGoroutineID()
	t.storage.Store(id, &valueWrapper{value: value})
	t.activeGoroutines.Store(id, struct{}{})
}

func (t *ThreadLocal) Get() (interface{}, bool) {
	id := getGoroutineID()
	if v, ok := t.storage.Load(id); ok {
		return v.(*valueWrapper).value, true
	}
	return nil, false
}

func (t *ThreadLocal) Remove() {
	id := getGoroutineID()
	t.storage.Delete(id)
	t.activeGoroutines.Delete(id)
}

func (t *ThreadLocal) cleanupRoutine() {
	for range t.cleanup.C {
		t.activeGoroutines.Range(func(key, _ interface{}) bool {
			id := key.(uint64)
			if !isGoroutineAlive(id) {
				t.storage.Delete(id)
				t.activeGoroutines.Delete(id)
			}
			return true
		})
	}
}

func (t *ThreadLocal) Stop() {
	t.cleanup.Stop()
}

func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(idField, 10, 64)
	return id
}

func isGoroutineAlive(id uint64) bool {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	return strings.Contains(string(buf[:n]), fmt.Sprintf("goroutine %d ", id))
}
