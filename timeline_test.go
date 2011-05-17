package main

import (
	. "launchpad.net/gocheck"
	"testing"
	"fmt"
	"bytes"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}
var _ = Suite(&S{})

func (s *S) TestBasicScheduling(c *C) {
	fmt.Printf("Sched\n");

	tl := NewTimeline(new(SimulatedTimer))
	b := bytes.NewBufferString("")

	tl.Schedule(101, func() {
		b.WriteString("c")
	})

	tl.Schedule(100, func() {
		b.WriteString("a")
	})
	
	tl.Schedule(100, func() {
		b.WriteString("b")
	})

	tl.Schedule(103, func() {
		b.WriteString("d")
	})
	
	hasMore := tl.ProcessNextEvent()
	if !hasMore { c.Fatal() }

	hasMore = tl.ProcessNextEvent()
	if !hasMore { c.Fatal() }

	hasMore = tl.ProcessNextEvent()
	if !hasMore { c.Fatal() }
	
	hasMore = tl.ProcessNextEvent()
	if hasMore { c.Fatal() }
	
	c.Assert("abcd", Equals, b.String())
}

func (s *S) TestSchedulePeriodic(c *C) {
	tl := NewTimeline(new(SimulatedTimer))
	b := bytes.NewBufferString("")

	tl.SchedulePeriodic(100, 10, 125, func() {
		b.WriteString("b")
	})

	hasMore := tl.ProcessNextEvent()
	if !hasMore { c.Fatal() }

	hasMore = tl.ProcessNextEvent()
	if !hasMore { c.Fatal() }
	
	hasMore = tl.ProcessNextEvent()
	if hasMore { c.Fatal() }
	
	c.Assert("bbb", Equals, b.String())
}
