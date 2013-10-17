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
	"regexp"
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

type ServiceGroup struct {
	Group string
	Services []*ServiceSnapshot
}

func (h *reqHandler) listServices(w http.ResponseWriter, r *http.Request) {
	ss := h.hub.GetServices()

	// group service by group
	groups := make(map[string] []*ServiceSnapshot)
	for _, s := range(ss) {
		groupServices, hasGroup := groups[s.Group]
		if !hasGroup {
			groupServices = make([]*ServiceSnapshot, 0, 100)
		}
		// copy because s gets overwritten each iteration
		t := s
		groups[s.Group] = append(groupServices, &t)
	}

	sg := make([]*ServiceGroup, 0, len(groups))
	for g, servicesForGroup := range(groups) {
		sg = append(sg, &ServiceGroup{g, servicesForGroup})
	}

	h.render("dashboard.tpl", map[string]interface{}{"groups":sg}, w)
}

func (h *reqHandler) listEventsData(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
//		log.Println("no service")
		return
	}

	startIndex := 0
	pageSize := 50

	startIndexStr, startIndexStrExists := r.Form["startIndex"]
	if startIndexStrExists {
		startIndex, _ = strconv.Atoi(startIndexStr[0])	
	}

	if pageSizeStr, pageSizeStrExists := r.Form["results"] ; pageSizeStrExists {
		pageSize, _ = strconv.Atoi(pageSizeStr[0])
	}

	entries := h.hub.GetLogEntries(serviceName[0])

	transformed := make([] map [string] interface{}, 0, 100)
	for index, l := range(entries) {
		if index < startIndex {
			// kinda stupid to loop through until we get to startIndex, but 
			// on a plane and no internet.  Can't look up the go docs at the
			// moment to see if I can give range a start/stop index.
			continue;
		}

		t := make(map[string] interface{})
		t["service"] = l.ServiceName
		t["summary"] = l.Summary
		t["severity"] = l.Severity
		t["timestamp"] = l.Timestamp
		t["id"] = l.Sequence

		transformed = append(transformed, t)

		// once we've got a full page, stop
		if len(transformed) >= pageSize {
			break;
		}
	}

	result := map[string]interface{}{"recordsReturned": len(transformed),
	    "totalRecords": len(entries),
	    "startIndex": startIndex,
	    "sort": nil,
	    "dir": nil,
	    "pageSize": pageSize,
	    "records": transformed }


//	log.Println("write result")

	enc := json.NewEncoder(w)
	enc.Encode(result)
}

func (h *reqHandler) showServiceStatus(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceNameArray, exists := r.Form["service"]
	if ! exists {
		return
	}
	serviceName := serviceNameArray[0]

	ss := h.hub.GetServices()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	result := "UNKNOWN"

	// group service by group
	for _, s := range(ss) {
		if s.Name == serviceName {
			if s.Enabled {
				result = "ON"
			} else {
				result = "OFF"
			}
			break
		}
	}

	w.Write([]byte(result))
}

func (h *reqHandler) listEvents(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceNameArray, exists := r.Form["service"]
	if ! exists {
		return
	}
	serviceName := serviceNameArray[0]

	filters := h.hub.GetNotificationFilters(serviceName)

	h.render("table.tpl", map[string]interface{}{"service":serviceName, "filters": filters}, w)
}

func (h *reqHandler) removeServiceEvents(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
		return
	}

	entries := h.hub.GetLogEntries(serviceName[0])
	for _, entry := range(entries) {
		h.hub.RemoveLogEntry(entry.Sequence)
	}
	http.Redirect(w, r, "/list-events", http.StatusTemporaryRedirect)
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

func (h *reqHandler) addNotificationFilter(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
		return
	}
	exprString, exprExists := r.Form["regexp"]
	if ! exprExists {
		return
	}

	regexp, expError := regexp.Compile(exprString[0])
	if expError != nil {
		log.Println("error: "+expError.String())
		return
	}

	h.hub.AddNotificationFilter(serviceName[0], regexp)

	http.Redirect(w, r, "/list-events?service="+serviceName[0], http.StatusTemporaryRedirect)
}

func (h *reqHandler) removeNotificationFilter(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	serviceName, exists := r.Form["service"]
	if ! exists {
		return
	}

	idString, idExists := r.Form["id"]
	if ! idExists {
		return
	}

	id, _ :=  strconv.Atoi(idString[0])

	h.hub.RemoveNotificationFilter(serviceName[0],  id)

	http.Redirect(w, r, "/list-events?service="+serviceName[0], http.StatusTemporaryRedirect)
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

func parseTimeOfDay(tstr string) int {
	parts := strings.Split(tstr, ":", 2)
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	return hour * 60 + minute
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

			group, groupPresent := conf.String(s, "group")
			if groupPresent != nil { 
				log.Println("group: "+groupPresent.String())
				group = "default"
			}

			description, descPresent := conf.String(s, "description")
			if descPresent != nil {
				log.Println("desc: "+descPresent.String())
				description = ""
			}

			enabled, enabledPresent := conf.Bool(s, "enabled")
			if enabledPresent != nil {
				log.Println("enabled: "+enabledPresent.String())
				enabled = false;
			}

			notificationStartTimeStr, nstartpresent := conf.String(s, "notificationsStart")
			if nstartpresent != nil {
				notificationStartTimeStr = "00:00"
			}
			notificationStart := parseTimeOfDay(notificationStartTimeStr)

			notificationStopTimeStr, nstoppresent := conf.String(s, "notificationsStop")
			if nstoppresent != nil {
				notificationStopTimeStr = "24:00"
			}
			notificationStop := parseTimeOfDay(notificationStopTimeStr)

			hub.AddService(name, heartbeatTimeout * 1000, group, description, enabled, notificationStart, notificationStop)
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
	http.HandleFunc("/service-status", func (w http.ResponseWriter, r *http.Request) {
		h.showServiceStatus(w, r)
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
	http.HandleFunc("/remove-service-events", func (w http.ResponseWriter, r *http.Request) {
		h.removeServiceEvents(w, r)
	})
	http.HandleFunc("/disable-service", func (w http.ResponseWriter, r *http.Request) {
		h.disableService(w, r)
	})
	http.HandleFunc("/enable-service", func (w http.ResponseWriter, r *http.Request) {
		h.enableService(w, r)
	})
	http.HandleFunc("/add-notification-filter", func (w http.ResponseWriter, r *http.Request) {
		h.addNotificationFilter(w, r)
	})
	http.HandleFunc("/remove-notification-filter", func (w http.ResponseWriter, r *http.Request) {
		h.removeNotificationFilter(w, r)
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
