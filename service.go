package main

import (
	"fmt"
	"bytes"
	"log"
	)

const ( 
	STATUS_UP = 1
	STATUS_DOWN = 2
	STATUS_UNKNOWN = 0
	)

const (
	OKAY = 0
	DEBUG = 1
	INFO = 2
	WARN = 3
	ERROR = 4
	)

type Service struct {
	Name string
	Monitor *HeartbeatMonitor
	Status int
	Log ServiceLog
}

type LogEntry struct {
	Summary string
	Severity int
	Timestamp int64
	Sequence int
}

type ServiceLog struct {
	entries []*LogEntry
}

type ServiceHub struct {
	timeline *Timeline
	services map[string] *Service
	notifier *Notifier
	logEntryCounter int
}

func NewServiceHub(timeline *Timeline) *ServiceHub {
	hub := &ServiceHub{timeline: timeline, services: make(map[string] *Service)}
	return hub
}

func (h *ServiceHub) AddService(serviceName string, heartbeatTimeout int) {
	heartbeatCallback := func(name string, isFailure bool) {
		if isFailure {
			h.Log(serviceName, WARN, "Heartbeat failure", h.timeline.Now())
		}
	}

	monitor := NewHeartbeatMonitor(h.timeline, serviceName, heartbeatTimeout, heartbeatCallback)
	s := &Service{Name: serviceName, Monitor: monitor, Status: STATUS_UNKNOWN}
	
	h.services[serviceName] = s
	
	monitor.Start()
}

func (h *ServiceHub) Heartbeat(serviceName string) {
	service := h.getService(serviceName)
	
	if service.Monitor != nil {
		log.Println("Calling heartbeat on \"%s\"", serviceName)
		service.Monitor.Heartbeat()
	} else {
		log.Println("Unknown service \"%s\"", serviceName)
	}
}

func (h *ServiceHub) getService(serviceName string) *Service {
	service, exists := h.services[serviceName]
	if !exists {
		service = &Service{Name: serviceName}
		h.services[serviceName] = service
	}
	return service
}

func (h *ServiceHub) Log(serviceName string, severity int, summary string, timestamp int64) {
	service := h.getService(serviceName)
	
	h.logEntryCounter += 1
	seq := h.logEntryCounter
	
	service.Log.entries = append(service.Log.entries, &LogEntry{summary, severity, timestamp, seq})

	h.notifier.CheckAndSendNotifications()
}

func (l *ServiceLog) FindAfter(sequence int) []*LogEntry {
	result := make([]*LogEntry, 0, len(l.entries))

	for _, v := range(l.entries) {
		if v.Sequence > sequence { 
			result = append(result, v)
		}
	}
	
	return result
}

const MAX_NOTIFY_FREQUENCY = 60

type ExecutorFn func (command string, input string)

type Notifier struct { 
	command string
	lastCheckSeq int
	lastSendTimestamp int64
	timeline *Timeline
	hub *ServiceHub
	executor ExecutorFn
}

func NewNotifier(command string, executor ExecutorFn, timeline *Timeline, hub *ServiceHub) *Notifier {
	return &Notifier{command: command, timeline: timeline, hub: hub, executor: executor}
}

func (n *Notifier) CheckAndSendNotifications() {
	now := n.timeline.Now()
	if now - n.lastSendTimestamp >= MAX_NOTIFY_FREQUENCY { 
		// enough time has passed since the last send 
		// so we can flush the event queue
		
		n.lastSendTimestamp = now
		n.sendNotificationSummary()
	} else {
		// too soon, so schedule a check of the queue after enough time has passed
		n.timeline.Schedule(n.lastSendTimestamp + MAX_NOTIFY_FREQUENCY, func() { n.CheckAndSendNotifications() } )
	}
}

func (n *Notifier) sendNotificationSummary() {
	msgsByService := make(map[string] []string)
	maxSeq := 0

	// find all outstanding notifications
	for k, v := range(n.hub.services) {
		e := v.Log.FindAfter(n.lastCheckSeq)
		if len(e) > 0 {
			msgs := make([]string, 0, len(e))
		
			for _, l := range(e) {
				if l.Sequence > maxSeq {
					maxSeq = l.Sequence
				}
				msgs = append(msgs, fmt.Sprintf("%s: %s", k, l.Summary))
			}

			msgsByService[k] = msgs
		}
	}
	
	if n.lastCheckSeq < maxSeq {
		n.lastCheckSeq = maxSeq
	}
	
	if len(msgsByService) > 1 {
		msg := bytes.NewBufferString("Multiple services had notifications: ")
		
		for k, v := range(msgsByService) {
			msg.WriteString(fmt.Sprintf("%s(%d) ", k, len(v)))
		}
	
		n.sendNotification(msg.String())
	} else if len(msgsByService) == 1 {
		// get the only msg list
		var serviceName string
		var msgs []string
		for tservice, tmsg := range(msgsByService) {
			serviceName = tservice
			msgs = tmsg
		}

		if len(msgs) > 1 { 
			// if we have multiple messages, just send the count of messages and 
			msg := fmt.Sprintf("%s had %d notifications", serviceName, len(msgs))
			n.sendNotification(msg)
		} else {
			// we must only have one message so just send that			
			msg := msgs[0]			
			n.sendNotification(msg)
		}
	}
	// otherwise if there were no messages pending, so do nothing
}

func (n *Notifier) sendNotification( msg string ) {
	n.executor(n.command, msg)
}
