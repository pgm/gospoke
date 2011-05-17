package main

import (
	"fmt"
	)

func main() {
	timeline := NewTimeline(new (RealTimer) )
	
	hub := NewServiceHub(timeline)
	notifier := NewNotifier("./notify_command.sh", ExecuteCommand, timeline, hub)

	hub.notifier = notifier

	hub.AddService("serviceA", 5 * 1000)
	hub.AddService("serviceB", 5 * 1000)

	fmt.Println("Starting rpc")
	go StartJsonRpcServer(hub)
	
	fmt.Println("Starting timeline")
	timeline.Run()
}