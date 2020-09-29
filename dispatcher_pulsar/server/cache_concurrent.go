package server

import (
	"errors"
	"sync"
	"time"
)

var (
	KeyNotFoundErr  = errors.New("KeyNotFoundErr")
	KeyAlreadyExist = errors.New("KeyAlreadyExist")
	SameValueErr    = errors.New("SameValueErr")
)

type DefaultSetFunc = func() (interface{}, error)

type Cache interface {
	Set(string, interface{}, time.Duration) error
	Get(string) (interface{}, error)
	Delete(string) error
	SetNX(string, interface{}, time.Duration) error
	Exists(string) bool
	SetDiff(string, string, time.Duration) error
}

type memoryCache struct {
	Cache
	values sync.Map
}

func (mc *memoryCache) Set(key string, val interface{}, expireDuration time.Duration) error {

	mc.values.Store(key, val)
	return nil
}

func (mc *memoryCache) Get(key string) (interface{}, error) {

	if val, ok := mc.values.Load(key); ok {
		return val, nil
	}
	return nil, KeyNotFoundErr
}

func (mc *memoryCache) Exists(key string) bool {

	_, ok := mc.values.Load(key)
	return ok
}

func (mc *memoryCache) SetNX(key string, val interface{}, expireDuration time.Duration) error {
	if _, ok := mc.values.LoadOrStore(key, val); ok {
		return KeyAlreadyExist
	} else {
		return nil
	}
}

func (mc *memoryCache) SetDiff(key string, val string, expireDuration time.Duration) error {

	old, ok := mc.values.Load(key)
	if ok && old.(string) == "finish" {
		return TaskAlreadyFinished
	} else if ok && old.(string) == val {
		return SameValueErr
	} else {
		mc.values.Store(key, val)
		return nil
	}
}

func (mc *memoryCache) Delete(key string) error {

	mc.values.Delete(key)
	return nil
}

func NewInMemoryCache() Cache {
	return &memoryCache{values: sync.Map{}}
}
