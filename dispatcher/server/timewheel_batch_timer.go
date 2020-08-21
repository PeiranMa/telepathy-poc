package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/RussellLuo/timingwheel"
)

type twBatchTimer struct {
	Timer
	tw     *timingwheel.TimingWheel
	mu     sync.Mutex
	items  map[string]*TimerItem
	timers map[string]*timingwheel.Timer
}

// func (t *twBatchTimer) tickItem(it *TimerItem) {
// 	needed := it.Tick()
// 	if time.Now().After(it.StartTime.Add(it.ExpireDuration)) {
// 		it.Timeout()
// 		goto exit
// 	}

// 	if needed {
// 		t.timers[it.ID] = t.tw.AfterFunc(it.TickDuration, func() {
// 			t.tickItem(it)
// 		})
// 	}
// exit: //motify
// 	t.Delete(it.ID)
// 	return
// }

func (t *twBatchTimer) Add(it *TimerItem) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.items[it.ID]; ok {
		return false
	}
	t.items[it.ID] = it
	it.StartTime = time.Now()
	timer := t.tw.AfterFunc(it.TickDuration, func() { it.Tick() })
	t.timers[it.ID] = timer
	return true
}

func (t *twBatchTimer) Delete(key string) {
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

func (t *twBatchTimer) Stop() {
	t.tw.Stop()
}

func NewTimingWheelBatch() *twBatchTimer {
	t := &twBatchTimer{
		items:  make(map[string]*TimerItem),
		timers: make(map[string]*timingwheel.Timer),
		tw:     timingwheel.NewTimingWheel(100*time.Millisecond, 100),
	}
	t.tw.Start()
	return t
}

func (t *twBatchTimer) GetItem(key string) *TimerItem {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.items[key]; !ok {
		fmt.Println("TimerItem not exit : ", key)
		return nil
	}
	return t.items[key]

}

// func (t *twBatchTimer) GetTimer(key string) *timingwheel.Timer {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()
// 	if _, ok := t.timers[key]; !ok {
// 		fmt.Println("TimerItem not exit.", key)
// 		return nil
// 	}
// 	return t.timers[key]

// }
func (t *twBatchTimer) SetTimer(key string, timer *timingwheel.Timer) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timers[key] = timer

}

func (t *twBatchTimer) AfterFunc(tickDuration time.Duration, tick func()) *timingwheel.Timer {
	return t.tw.AfterFunc(tickDuration, tick)
}
