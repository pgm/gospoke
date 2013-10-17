package main

import (
	"container/heap"
	"bytes"
	"fmt"
	"time"
	"sync"
	)

type CallbackFn func()

type Event struct {
	Timestamp time.Time
	InsertSeq int
	Callback CallbackFn
} 

type EventSlice struct {
	events []*Event
}

type Timer interface {
	Now() time.Time
	SleepUntil(cond *sync.Cond, time time.Time)
	Sleep(cond *sync.Cond)
}

type Timeline struct {
	timer Timer
	nextSeq int
	events *EventSlice

	cond *sync.Cond
	lock sync.Locker
}

func NewTimeline(timer Timer) *Timeline {
	t := new(Timeline)
	t.timer = timer
	t.events = new (EventSlice)

	t.lock = new(sync.Mutex)
	t.cond = sync.NewCond(t.lock)
	
	return t
}

func (v EventSlice) Len() int {
	return len(v.events)
}

func (v EventSlice) Less(i, j int) bool {
	tdelta := v.events[i].Timestamp.Sub(v.events[j].Timestamp)
	if tdelta == 0 {
		return v.events[i].InsertSeq < v.events[j].InsertSeq
	}
	return tdelta < 0
}

func (v EventSlice) Swap(i, j int) { 
	v.events[i], v.events[j] = v.events[j], v.events[i]
}

func (v *EventSlice) Push(x interface{} ) {
	e := x.(*Event)
	v.events = append(v.events, e)
}

func (v *EventSlice) Pop() interface{} {
	x := v.events[len(v.events)-1]
	v.events = v.events[:len(v.events)-1]
	return x
}

func (v EventSlice) String() string {
	b := new(bytes.Buffer)
	b.WriteString("[")
	for i, e := range v.events {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(fmt.Sprintf("%s", e))
	}
	b.WriteString("]")

	return b.String()
}

func (t Timeline) Now () time.Time {
	return t.timer.Now()
}

func (t *Timeline) Schedule(timestamp time.Time, callback CallbackFn) {
	t.lock.Lock()
	defer t.lock.Unlock()
	
	seq := t.nextSeq
	t.nextSeq += 1
	heap.Push(t.events, &Event{timestamp, seq, callback})
	t.cond.Broadcast()
}

func (t *Timeline) SchedulePeriodic(timestamp time.Time, period time.Duration, endTime time.Time, callback CallbackFn) {
	nextTime := timestamp 
	var execAndReschedule CallbackFn
	execAndReschedule = func() {
		nextTime = nextTime.Add(period)
		if nextTime.Before(endTime) {
			t.Schedule(nextTime, execAndReschedule)
		}
		callback()
	}
	t.Schedule(nextTime, execAndReschedule)
}

func (t *Timeline) peek() *Event {
	if t.events.Len() > 0 {
		return t.events.events[0]
	} 

	return nil
}

func (t *Timeline) processNextEventAssumingLocked() bool {
	for {
		e := t.peek()
		if e == nil {
			break
		} else {
		
			if t.timer.Now().Before(e.Timestamp) {
				// if we haven't reached the 
				// timestamp of the next time then 
				// go sleep until that time.  Now 
				// it's possible while we're sleeping 
				// that a new event could have gotten
				// equeued.  So when we wake up, peek
				// the top event again from scratch

				t.timer.SleepUntil(t.cond, e.Timestamp)
				
			} else {
				removed := heap.Pop(t.events)
				if(e != removed) { panic("wrong element removed") }

				t.lock.Unlock()
				e.Callback()
				t.lock.Lock()
				
				break
			}
		}
	}
	
	return t.events.Len() > 0
}

func (t *Timeline) ProcessNextEvent() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.processNextEventAssumingLocked()
}

func (t *Timeline) Run() {
	t.lock.Lock()
	defer t.lock.Unlock()
	
	for {
		hasMore := 	t.processNextEventAssumingLocked()
		// race condition here
		if !hasMore {
			t.timer.Sleep(t.cond)
		}
	}
}

func (t *Timeline) sleepUntilNextEventOrTimestamp(timestamp time.Time) bool {
	for {
		e := t.peek()
		if e == nil || e.Timestamp.After(timestamp) {
			t.timer.SleepUntil(t.cond, timestamp)
			
			if !(t.Now().Before(timestamp)) {
				return false
			}
		} else {
			break
		}
	}
	
	return true
}

func (t *Timeline) RunUntil(timestamp time.Time) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for {
		if !t.sleepUntilNextEventOrTimestamp(timestamp) {
			break
		}
		t.processNextEventAssumingLocked()
	}
}

func (t *Timeline) Execute(c CallbackFn) {
	t.Schedule(t.Now(), c)
}


////////////

type SimulatedTimer struct { 
	time time.Time
}

func (t SimulatedTimer) Now() time.Time {
	return t.time
}

func (t *SimulatedTimer) SleepUntil(cond *sync.Cond, time time.Time) {
	if t.time.Before(time) {
		t.time = time
	}
}

func (t SimulatedTimer) Sleep(cond *sync.Cond) {
	panic("Cannot sleep indefinitely with a simulated timer")
}

////////////

type RealTimer struct {
}

func (t RealTimer) Now() time.Time {
	now := time.Now()
//	if err != nil {
//		log.Fatal("Could not get time: %v", err)
//	}

	return now
}

func (t *RealTimer) SleepUntil(cond *sync.Cond, timestamp time.Time) {
	now := t.Now()
	delay := timestamp.Sub(now)
	if delay > 0 {
		tt := time.AfterFunc(time.Duration(delay), func() { cond.Broadcast() } );
		t.Sleep(cond)
		tt.Stop()
	}
}

func (t RealTimer) Sleep(cond *sync.Cond) {
	cond.Wait()
}


