package main

import (
	"log"
	)

type HeartbeatFailureCallback func(name string, isFailure bool) 

type HeartbeatMonitor struct {
	timeline *Timeline
	name string
	period int
	lastHeartbeat int64
	callback HeartbeatFailureCallback
	failed bool
}

func (h *HeartbeatMonitor) scheduleHeartbeatTimeout() {
	nextTimeout := h.lastHeartbeat + int64(h.period)
	log.Printf("scheduling timeout for %v\n", nextTimeout);
	h.timeline.Schedule(nextTimeout, func() { h.checkHeartbeatTimeout() } )
}

func (h *HeartbeatMonitor) checkHeartbeatTimeout() {
	if h.timeline.Now() - h.lastHeartbeat >= int64(h.period) {
		log.Println("failure",h.name);
		h.failed = true
		h.callback(h.name, h.failed)
	}
}

func (h *HeartbeatMonitor) Start() {
	h.Heartbeat()
}

func (h *HeartbeatMonitor) Heartbeat() {
	h.lastHeartbeat = h.timeline.Now()
	if h.failed {
		log.Println("okay",h.name);
		h.failed = false
		h.callback(h.name, h.failed)
	}
	h.scheduleHeartbeatTimeout()	
}

func NewHeartbeatMonitor ( timeline *Timeline, name string, period int, callback HeartbeatFailureCallback) *HeartbeatMonitor {
	m := &HeartbeatMonitor{timeline, name, period, 0, callback, false}
	return m
}
