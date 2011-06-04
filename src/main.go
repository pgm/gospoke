package main

import (
	"log"
	"http"
	"os"
	"github.com/kless/goconfig/config"
	"net"
	"strings"
	"github.com/hoisie/mustache.go"
	"path"
	"json"
	"strconv"
	"flag"
	)

type reqHandler struct {
	hub ThreadSafeServiceHub
	templateDir string
}

func (h *reqHandler) render(filename string, context interface{}, w http.ResponseWriter) {
	t, err := mustache.ParseFile(h.templateDir + "/" + filename)
	
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		log.Println(err.String())
		return
	}

	result := t.Render(context)
	w.Write([]byte(result))
}

func (h *reqHandler) listServices(w http.ResponseWriter, r *http.Request) {
	ss := h.hub.GetServices()

	h.render("dashboard.tpl", map[string]interface{}{"services":ss}, w)
}

func (h *reqHandler) listEventsData(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
		return
	}

	entries := h.hub.GetLogEntries(serviceName[0])

	transformed := make([] map [string] interface{}, 0, 100)
	for _, l := range(entries) {
		t := make(map[string] interface{})
		t["service"] = l.ServiceName
		t["summary"] = l.Summary
		t["severity"] = l.Severity
		t["timestamp"] = l.Timestamp
		t["id"] = l.Sequence

		transformed = append(transformed, t)
	}

	result := map[string]interface{}{"recordsReturned": len(transformed),
	    "totalRecords": len(transformed),
    	"startIndex": 0,
      	"sort": nil,
     	"dir": nil,
      	"pageSize": 10,
      	"records": transformed }

	enc := json.NewEncoder(w)
	enc.Encode(result)
}

func (h *reqHandler) listEvents(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceNameArray, exists := r.Form["service"]
	if ! exists {
		return
	}
	serviceName := serviceNameArray[0]

	h.render("table.tpl", map[string]interface{}{"service":serviceName}, w)
}

func (h *reqHandler) removeEvents(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	eventIds, exists := r.Form["id"]
	if exists {
		for _, eventIdStr := range(eventIds) {
			eventId, err := strconv.Atoi(eventIdStr)
			if ( err == nil ) {	
				h.hub.RemoveLogEntry(eventId)
			}
		}
	}
	http.Redirect(w, r, "/list-events", http.StatusTemporaryRedirect)
}

func (h *reqHandler) disableService(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
		return
	}
	h.hub.SetServiceEnabled(serviceName[0], false)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *reqHandler) enableService(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
		return
	}
	h.hub.SetServiceEnabled(serviceName[0], true)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *reqHandler) makeFileServer(directory string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, urlFilename := path.Split(r.RawURL)

		filename := path.Join(directory, urlFilename)

		fi, err := os.Stat(filename)

		if err == nil && fi.IsRegular() {
			http.ServeFile(w, r, filename)
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	
	configFilename := "gospoke.ini"
	if len(args) > 0 {
		configFilename = args[0]
	}
	
	conf, err := config.ReadDefault(configFilename)
	
	if err != nil {
		log.Fatalln(err)
	}

	timeline := NewTimeline(new (RealTimer) )
	hub := NewServiceHub(timeline)

	notifierCommand, _ := conf.String("default", "notifier_command")
	notifierThrottle, _ := conf.Int("default", "notifier_throttle")
	listeningAddr, _ := conf.String("default", "listen")
	resourceDir, _ := conf.String("default", "resource_dir")

	notifier := NewNotifier(notifierCommand, notifierThrottle * 1000, ExecuteCommand, timeline, hub)
	
	for _, s := range(conf.Sections()) {
		if strings.HasPrefix(s, "service") {
			name, _ := conf.String(s, "name")
			heartbeatTimeout, _ := conf.Int(s, "timeout")

			hub.AddService(name, heartbeatTimeout * 1000)
		}
	}

	hub.notifier = notifier
	threadSafeHub := NewHubAdapter(hub)

	h := &reqHandler{threadSafeHub, resourceDir+"/views"}

	http.Handle("/jsonrpc", MewJsonRpcHandler(threadSafeHub, timeline))
	http.HandleFunc("/blueprint/", h.makeFileServer(resourceDir+"/css/blueprint"))
	http.HandleFunc("/css/", h.makeFileServer(resourceDir+"/css"))
	http.HandleFunc("/img/", h.makeFileServer(resourceDir+"/img"))
	http.HandleFunc("/js/", h.makeFileServer(resourceDir+"/js"))

	// I've got enough of these maybe I should refactor somehow
	// punting because perhaps I can find an existing framework to adopt instead
	// of rolling my own.
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		h.listServices(w, r)
	})
	http.HandleFunc("/list-events", func (w http.ResponseWriter, r *http.Request) {
		h.listEvents(w, r)
	})
	http.HandleFunc("/list-events-data", func (w http.ResponseWriter, r *http.Request) {
		h.listEventsData(w, r)
	})
	http.HandleFunc("/remove-events", func (w http.ResponseWriter, r *http.Request) {
		h.removeEvents(w, r)
	})
	http.HandleFunc("/disable-service", func (w http.ResponseWriter, r *http.Request) {
		h.disableService(w, r)
	})
	http.HandleFunc("/enable-service", func (w http.ResponseWriter, r *http.Request) {
		h.enableService(w, r)
	})
	
	log.Println("Starting http server on "+listeningAddr)

	l, e := net.Listen("tcp", listeningAddr)
	if e != nil {
		log.Fatalln(e)
	}
	go http.Serve(l, nil)

	log.Println("Starting timeline")
	timeline.Run()
}
