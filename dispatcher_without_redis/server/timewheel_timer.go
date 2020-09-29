package server

import (
	"sync"
	"time"

	"github.com/RussellLuo/timingwheel"
)

type TimerItem struct {
	Timeout        func()
	ExpireDuration time.Duration
	ID             string
}
type Timer interface {
	Add(*TimerItem) bool
	Delete(string)
	Stop()
}

type twTimer struct {
	tw     *timingwheel.TimingWheel
	mu     sync.Mutex
	items  map[string]*TimerItem
	timers map[string]*timingwheel.Timer
}

func (t *twTimer) Add(it *TimerItem) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.items[it.ID]; ok {
		return false
	}
	t.items[it.ID] = it
	timer := t.tw.AfterFunc(it.ExpireDuration, it.Timeout)
	t.timers[it.ID] = timer
	return true
}

func (t *twTimer) Delete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.items[key]; !ok {
		return
	}
	delete(t.items, key)
	timer := t.timers[key]
	timer.Stop()
	delete(t.timers, key)
}

func (t *twTimer) Stop() {
	t.tw.Stop()
}

func NewTimingWheel() Timer {
	t := &twTimer{
		items:  make(map[string]*TimerItem),
		timers: make(map[string]*timingwheel.Timer),
		tw:     timingwheel.NewTimingWheel(100*time.Millisecond, 100),
	}
	t.tw.Start()
	return t
}
