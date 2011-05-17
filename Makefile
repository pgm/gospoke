include $(GOROOT)/src/Make.inc

TARG=mypackage

GOFILES=\
	timeline.go\
	heartbeat.go\
	endpoint.go\
	service.go\
	executor.go\
	main.go

include $(GOROOT)/src/Make.cmd
