package main

import (
	"log"
	"http"
	"template"
	"io"
	"github.com/kless/goconfig/config"
	"fmt"
	"strings"
	)

type snapshots struct {
	Snapshots []ServiceSnapshot
}

func listServices(hub* ServiceHub, w http.ResponseWriter, r *http.Request) {
	fmtrs := make(template.FormatterMap)
	fmtrs["status"] = func(w io.Writer, formatter string, data ...interface{}) {
		v := data[0].(int)
		s := "unknown"
		
		if v == STATUS_UP {
			s = "up"
		} else if v == STATUS_DOWN {
			s = "down"
		}
		
		w.Write([]uint8(s))
	}

	t, err := template.ParseFile("index.html", fmtrs)
	
	if err != nil {
		//http.Redirect(w, r, "/", http.StatusFound)		
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	

	ss := hub.GetServiceSnapshots()
	err = t.Execute(w, &snapshots{ss})
	if err != nil {
		log.Println(err.String())
	}
}

func main() {
	conf, err := config.ReadDefault("gospoke.ini")
	
	if err != nil {
		fmt.Println(err)
	}

	timeline := NewTimeline(new (RealTimer) )
	hub := NewServiceHub(timeline)

	notifierCommand, _ := conf.String("default", "notifier_command")
	notifierThrottle, _ := conf.Int("default", "notifier_throttle")

	notifier := NewNotifier(notifierCommand, notifierThrottle * 1000, ExecuteCommand, timeline, hub)
	
	for _, s := range(conf.Sections()) {
		if strings.HasPrefix(s, "service") {
			name, _ := conf.String(s, "name")
			heartbeatTimeout, _ := conf.Int(s, "timeout")

			hub.AddService(name, heartbeatTimeout * 1000)
		}
	}

	hub.notifier = notifier

	log.Println("Starting rpc")
	go StartJsonRpcServer(hub)
	
	log.Println("Starting timeline")
	go timeline.Run()

	log.Println("Starting http server")

	http.HandleFunc("/css/", func (w http.ResponseWriter, r *http.Request) { 
		http.ServeFile(w, r, filename)
	})
	
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		listServices(hub, w, r)
	})
	
	http.ListenAndServe(":8080", nil)
}
