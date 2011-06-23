package main

import (
	"fmt"
	"bytes"
//	"log"
	"time"
	"sort"
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
	Enabled bool
	Monitor *HeartbeatMonitor
	Status int
	HeartbeatCount int
	LastHeartbeatTimestamp int64
	Log ServiceLog
}

type LogEntry struct {
	ServiceName string
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

type ServiceSnapshot struct {
	Name string
	Status int
	LastHeartbeatTimestamp string
	IsUp bool
	IsDown bool
	IsUnknown bool
	Enabled bool
	Notifications []NotificationSummary
}

type NotificationSummary struct {
	Severity int
	Count int
}

type ApiError struct {
	error string
}

func (e ApiError) String() string {
	return e.error
}

// contract between service hub and all threads running outside of 
// timeline thread
type ThreadSafeServiceHub interface {
	Log(serviceName string, summary string, severity int, timestamp int64) *ApiError
	Heartbeat(serviceName string) *ApiError

	GetLogEntries(serviceName string) []*LogEntry
	RemoveLogEntry(sequence int)
	GetServices() []ServiceSnapshot

	SetServiceEnabled(serviceName string, enabled bool)
}

type ServiceHubAdapter struct {
	hub *ServiceHub
}

func (a *ServiceHubAdapter) SetServiceEnabled(serviceName string, enabled bool) {
	c := make(chan *ApiError)
	hub := a.hub

	hub.timeline.Execute(func() {
		c <- hub.SetServiceEnabled(serviceName, enabled)
	})

	<-c
}

func (a *ServiceHubAdapter)	Log(serviceName string, summary string, severity int, timestamp int64) *ApiError {
	c := make(chan *ApiError)
	hub := a.hub

	hub.timeline.Execute(func() {
		c <- hub.Log(serviceName, summary, severity, timestamp)
	})

	return <-c
}

func (a *ServiceHubAdapter) Heartbeat (serviceName string) *ApiError  {
	c := make(chan *ApiError)
	hub := a.hub

	hub.timeline.Execute(func() {
		service, found := hub.services[serviceName]

		if !found {
			c <- &ApiError{"No service named \""+serviceName+"\""}
			return 
		}

		if service.Monitor != nil {
			service.Monitor.Heartbeat()
		}

		c <- nil
	})

	return <-c
}

func (a *ServiceHubAdapter) GetServices() []ServiceSnapshot {
	c := make(chan []ServiceSnapshot)
	hub := a.hub

	hub.timeline.Execute(func() {
		ss := make([]ServiceSnapshot, 0, len(hub.services))
		
		for _, v := range(hub.services) {
			notifications := make([]NotificationSummary, 0, 10)

			// count the number of message per severity
			counts := make(map[int] int)
			for _, l := range(v.Log.entries) {
				c, exists := counts[l.Severity]
				if !exists {
					c = 0
				}
				c += 1
				counts[l.Severity] = c
			}

			// now add them to the notification list ordered by severity
			keys := make([]int, 0, len(notifications))
			for k, _ := range(counts) {
				keys = append(keys, k)
			}
			sort.SortInts(keys)

			for _, k := range(keys) {
				notifications = append(notifications, NotificationSummary{k, counts[k]})
			}

			var timestamp string
			if v.HeartbeatCount == 0 {
				timestamp = ""
			} else {
				timestamp = time.SecondsToLocalTime(v.LastHeartbeatTimestamp/1000).Format(time.Kitchen)
			}

			ss = append(ss, ServiceSnapshot{v.Name, v.Status, timestamp, v.Status == STATUS_UP, v.Status == STATUS_DOWN, v.Status == STATUS_UNKNOWN, v.Enabled, notifications})
		}
		c <- ss		
	})

	return <-c
}

func (a *ServiceHubAdapter) GetLogEntries(serviceName string) []*LogEntry {
	c := make(chan []*LogEntry)
	hub := a.hub

	hub.timeline.Execute(func() {
		ss := make([]*LogEntry, 0, 100)

		service, found := hub.services[serviceName]
		if ! found {
			c <- ss
			return
		}

		for _, v := range(service.Log.entries) {
			ss = append(ss, v)
		}

		c <- ss
	})

	return <-c
}

func removeLogEntriesWithId(entries []*LogEntry, sequenceToDel int) []*LogEntry {
	dest := 0
	for i, v := range(entries) {
		if v.Sequence == sequenceToDel { 
			continue
		}
		entries[dest] = entries[i]
		dest++
	}
	return entries[:dest]
}

func (a *ServiceHubAdapter) RemoveLogEntry(sequence int) {
	c := make(chan bool)
	hub := a.hub

	hub.timeline.Execute(func() {
		for _, service := range(hub.services) {
			service.Log.entries = removeLogEntriesWithId(service.Log.entries, sequence)
		}

		c <- true
	})
		
	<-c
}

func NewHubAdapter(hub *ServiceHub) *ServiceHubAdapter {
	return &ServiceHubAdapter{hub}
}

////////////////////////////////////////////////////////////////////////

func NewServiceHub(timeline *Timeline) *ServiceHub {
	hub := &ServiceHub{timeline: timeline, services: make(map[string] *Service)}
	hub.logEntryCounter = 1
	return hub
}

func (h *ServiceHub) SetServiceEnabled(serviceName string, enabled bool) *ApiError {
	service, found := h.services[serviceName]

	if !found {
		return &ApiError{"No service named \""+serviceName+"\""}
	}

	service.Enabled = enabled	

	return nil
}

func (h *ServiceHub) Log(serviceName string, summary string, severity int, timestamp int64) *ApiError {
	service, found := h.services[serviceName]

	if !found {
		return &ApiError{"No service named \""+serviceName+"\""}
	}

	seq := h.nextSequenceId()
	service.Log.entries = append(service.Log.entries, &LogEntry{serviceName, summary, severity, timestamp, seq})
	h.notifier.CheckAndSendNotifications()

	return nil
}

func (h *ServiceHub) AddService(serviceName string, heartbeatTimeout int) {
	var s *Service
	s = &Service{Name: serviceName, Enabled: false, Status: STATUS_UNKNOWN}

	heartbeatCallback := func(name string, isFailure bool) {
		if isFailure {
			h.Log(serviceName, "Heartbeat failure", WARN,  h.timeline.Now())
			s.Status = STATUS_DOWN
		} else {
			s.Status = STATUS_UP
			s.HeartbeatCount += 1
			s.LastHeartbeatTimestamp = h.timeline.Now()
		}
	}

	s.Monitor = NewHeartbeatMonitor(h.timeline, serviceName, heartbeatTimeout, heartbeatCallback)
	
	h.services[serviceName] = s
	
	s.Monitor.Start()
}


func (h *ServiceHub) nextSequenceId() int {
	h.logEntryCounter += 1
	seq := h.logEntryCounter
	return seq
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

type ExecutorFn func (command string, input string)

type Notifier struct { 
	command string
	lastCheckSeq int
	lastSendTimestamp int64
	timeline *Timeline
	hub *ServiceHub
	executor ExecutorFn
	throttle int64
}

func NewNotifier(command string, throttle int, executor ExecutorFn, timeline *Timeline, hub *ServiceHub) *Notifier {
	return &Notifier{command: command, throttle: int64(throttle), timeline: timeline, hub: hub, executor: executor}
}

func (n *Notifier) CheckAndSendNotifications() {
	now := n.timeline.Now()
	if now - n.lastSendTimestamp >= n.throttle { 
		// enough time has passed since the last send 
		// so we can flush the event queue
		
		n.lastSendTimestamp = now
		n.sendNotificationSummary()
	} else {
		// too soon, so schedule a check of the queue after enough time has passed
		n.timeline.Schedule(n.lastSendTimestamp + n.throttle, func() { n.CheckAndSendNotifications() } )
	}
}

func (n *Notifier) sendNotificationSummary() {

	// find all outstanding notifications, grouping them by service
	msgsByService := make(map[string] []string)
	maxSeq := 0
	for k, v := range(n.hub.services) {
		if !v.Enabled{
			continue
		}

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
