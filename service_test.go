package main

import (
	. "launchpad.net/gocheck"
	"bytes"
	"fmt"
)

func SetupNotifier() (result *bytes.Buffer, tl *Timeline, hub *ServiceHub) {
	result = bytes.NewBufferString("")
	
	tl = NewTimeline(new(SimulatedTimer))
	
	hub = NewServiceHub(tl)
	
	n := new(Notifier)
	n.command = "cmd"
	n.timeline = tl
	n.hub = hub
	n.executor = func(cmd string, input string) {
		now := tl.Now()
		t := fmt.Sprintf("%d:%s(%s)", now, cmd, input)
		result.WriteString( t )
	}
	hub.notifier = n

	return
}

func (s *S) TestSingleNotification(c *C) {
	result, tl, hub := SetupNotifier()
	
	tl.Schedule(100, func() { hub.Log("service", WARN, "warn", 0) } )
	tl.RunUntil(1000)
	
	c.Assert(result.String(), Equals, "100:cmd(service: warn)")
}

func (s *S) TestTwoQuickNotification(c *C) {
	result, tl, hub := SetupNotifier()
	
	tl.Schedule(100, func() { hub.Log("service", WARN, "warn1", 0) } )
	tl.Schedule(101, func() { hub.Log("service", WARN, "warn2", 0) } )
	tl.RunUntil(1000)
	
	c.Assert(result.String(), Equals, "100:cmd(service: warn1)160:cmd(service: warn2)")
}

func (s *S) TestManyQuickNotification(c *C) {
	result, tl, hub := SetupNotifier()
	
	tl.Schedule(100, func() { hub.Log("service", WARN, "warn1", 0) } )
	tl.Schedule(101, func() { hub.Log("service", WARN, "warn2", 0) } )
	tl.Schedule(102, func() { hub.Log("service", WARN, "warn3", 0) } )
	tl.Schedule(103, func() { hub.Log("service", WARN, "warn4", 0) } )
	tl.RunUntil(1000)
	
	c.Assert(result.String(), Equals, "100:cmd(service: warn1)160:cmd(service had 3 notifications)")
}

func (s *S) TestManyQuickNotificationDifferentServices(c *C) {
	result, tl, hub := SetupNotifier()
	
	tl.Schedule(100, func() { hub.Log("service", WARN, "warn1", 0) } )
	tl.Schedule(101, func() { hub.Log("alt1", WARN, "warn2", 0) } )
	tl.Schedule(102, func() { hub.Log("alt2", WARN, "warn3", 0) } )
	tl.Schedule(103, func() { hub.Log("alt3", WARN, "warn4", 0) } )
	tl.RunUntil(1000)
	
	c.Assert(result.String(), Equals, "100:cmd(service: warn1)160:cmd(Multiple services had notifications: alt3(1) alt2(1) alt1(1) )")
}
