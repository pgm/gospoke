package main

import (
	. "launchpad.net/gocheck"
	"bytes"
)


func (s *S) TestHeartbeatFailure(c *C) {
	b := bytes.NewBufferString("")
	callback := func(name string, failed bool) { 
		b.WriteString(name)
	}

	tl := NewTimeline(new(SimulatedTimer))
	m := NewHeartbeatMonitor(tl, "n", 20, callback)
	m.Start()

	tl.RunUntil(1000)
	
	c.Assert(b.String(), Equals, "n")
}

func (s *S) TestHeartbeatSuccess(c *C) {
	b := bytes.NewBufferString("")
	callback := func(name string, failed bool) { 
		b.WriteString(name)
		if failed {
			b.WriteString("f")
		} else {
			b.WriteString("c")
		}
	}

	tl := NewTimeline(new(SimulatedTimer))
	m := NewHeartbeatMonitor(tl, "n", 20, callback)
	
	tl.Schedule(100, func() { m.Heartbeat() })
	
	m.Start()

	tl.RunUntil(1000)
	
	c.Assert(b.String(), Equals, "nfncnf")
}