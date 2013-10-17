package main

import (
	"log"
	"time"
	)

type HeartbeatFailureCallback func(name string, isFailure bool) 

type HeartbeatMonitor struct {
	timeline *Timeline
	name string
	period time.Duration
	lastHeartbeat time.Time
	callback HeartbeatFailureCallback
	failed bool
}

func (h *HeartbeatMonitor) scheduleHeartbeatTimeout() {
	nextTimeout := h.lastHeartbeat.Add(h.period)
	h.timeline.Schedule(nextTimeout, func() { h.checkHeartbeatTimeout() } )
}

func (h *HeartbeatMonitor) checkHeartbeatTimeout() {
	if h.timeline.Now().Sub(h.lastHeartbeat) >= h.period {
		log.Println("failure",h.name);
		h.failed = true
		h.callback(h.name, h.failed)
	}
}

func (h *HeartbeatMonitor) Start() {
	h.lastHeartbeat = h.timeline.Now()
	h.scheduleHeartbeatTimeout()
}

func (h *HeartbeatMonitor) Heartbeat() {
	h.lastHeartbeat = h.timeline.Now()
	if h.failed {
		log.Println("okay",h.name);
		h.failed = false
	} 
	h.callback(h.name, h.failed)
	h.scheduleHeartbeatTimeout()	
}

func NewHeartbeatMonitor ( timeline *Timeline, name string, period time.Duration, callback HeartbeatFailureCallback) *HeartbeatMonitor {
	m := &HeartbeatMonitor{timeline, name, period, time.Now(), callback, false}
	return m
}
