package server

import (
	// "errors"
	"sync"
	"time"
)

var (
// KeyNotFoundErr            = errors.New("KeyNotFoundErr")
// KeyAlreadyExist           = errors.New("KeyAlreadyExist")
// TaskAlreadyInWorkingState = errors.New("TaskAlreadyInWorkingState")
// TaskAlreadyFinished       = errors.New("TaskAlreadyFinished")
// SameValueErr = errors.New("SameValueErr")
)

type memoryStateCache struct {
	sync.Mutex
	Cache
	values map[string]string
}

func (mc *memoryStateCache) Set(key string, val string, expireDuration time.Duration) error {
	mc.Lock()
	defer mc.Unlock()
	mc.values[key] = val
	return nil
}

func (mc *memoryStateCache) Get(key string) (string, error) {
	mc.Lock()
	defer mc.Unlock()
	if val, ok := mc.values[key]; ok {
		return val, nil
	}
	return "", KeyNotFoundErr
}

func (mc *memoryStateCache) Exists(key string) bool {
	mc.Lock()
	defer mc.Unlock()
	_, ok := mc.values[key]
	return ok
}

func (mc *memoryStateCache) SetNX(key string, val string, expireDuration time.Duration) error {
	mc.Lock()
	defer mc.Unlock()
	if _, ok := mc.values[key]; ok {
		return KeyAlreadyExist
	} else {
		mc.values[key] = val
		return nil
	}
}

func (mc *memoryStateCache) SetDiff(key string, val string, expireDuration time.Duration) error {
	mc.Lock()
	defer mc.Unlock()
	old, ok := mc.values[key]
	if ok && old == "finish" {
		return TaskAlreadyFinished
	} else if ok && old == val {
		return SameValueErr

	} else {
		mc.values[key] = val
		return nil
	}
}

func (mc *memoryStateCache) Delete(key string) error {
	mc.Lock()
	defer mc.Unlock()
	delete(mc.values, key)
	return nil
}

func NewInmemoryStateCache() *memoryStateCache {
	return &memoryStateCache{values: make(map[string]string)}
}
