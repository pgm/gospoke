package main

import (
	"container/heap"
	"bytes"
	"fmt"
	"os"
	"time"
	"log"
	"sync"
	)

type CallbackFn func()

type Event struct {
	Timestamp int64
	InsertSeq int
	Callback CallbackFn
} 

type EventSlice struct {
	events []*Event
}

type Timer interface {
	Now() int64
	SleepUntil(cond *sync.Cond, time int64)
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
	tdelta := v.events[i].Timestamp - v.events[j].Timestamp
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

func (t Timeline) Now () int64 {
	return t.timer.Now()
}

func (t *Timeline) Schedule(timestamp int64, callback CallbackFn) {
	log.Println("s lock")
	t.lock.Lock()
	log.Println("s locked")

	defer t.lock.Unlock()
	
	seq := t.nextSeq
	t.nextSeq += 1
	heap.Push(t.events, &Event{timestamp, seq, callback})
	t.cond.Broadcast()
}

func (t *Timeline) SchedulePeriodic(timestamp int64, period int, endTime int64, callback CallbackFn) {
	nextTime := timestamp 
	var execAndReschedule CallbackFn
	execAndReschedule = func() {
		nextTime += int64(period)
		if nextTime < endTime {
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
		
			if t.timer.Now() < e.Timestamp {
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
				log.Println("unlock")
				e.Callback()
				log.Println("lock")
				t.lock.Lock()
				log.Println("locked")
				
				break
			}
		}
	}
	
	return t.events.Len() > 0
}

func (t *Timeline) ProcessNextEvent() bool {
	log.Println("p lock")
	t.lock.Lock()
	log.Println("p locked")
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

func (t *Timeline) sleepUntilNextEventOrTimestamp(timestamp int64) bool {
	for {
		e := t.peek()
		if e == nil || e.Timestamp > timestamp {
			t.timer.SleepUntil(t.cond, timestamp)
			
			if t.Now() >= timestamp {
				return false
			}
		} else {
			break
		}
	}
	
	return true
}

func (t *Timeline) RunUntil(timestamp int64) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for {
		if !t.sleepUntilNextEventOrTimestamp(timestamp) {
			break
		}
		t.processNextEventAssumingLocked()
	}
}


////////////

type SimulatedTimer struct { 
	time int64
}

func (t SimulatedTimer) Now() int64 {
	return t.time
}

func (t *SimulatedTimer) SleepUntil(cond *sync.Cond, time int64) {
	if t.time < time {
		t.time = time
	}
	log.Println("Sleeping until", time)
}

func (t SimulatedTimer) Sleep(cond *sync.Cond) {
	panic("Cannot sleep indefinitely with a simulated timer")
}

////////////

type RealTimer struct {
}

func (t RealTimer) Now() int64 {
	sec, nsec, err := os.Time()
	if err != nil {
		log.Fatal("Could not get time: %v", err)
	}

	return sec * 1000 + nsec/1000000
}

func (t *RealTimer) SleepUntil(cond *sync.Cond, timestamp int64) {
	now := t.Now()
	delay := timestamp - now
	if delay > 0 {
		go func() {
			time.Sleep(delay*1000000)
			cond.Broadcast()
		}()
		t.Sleep(cond)
	}
}

func (t RealTimer) Sleep(cond *sync.Cond) {
	cond.Wait()
}


