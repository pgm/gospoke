package main

import (
	"http"
//	"log"
	"json"
	"io"
	)

type Sample struct { }

type Args struct { }

type JsonRpcCallback func (map[string] interface{}) interface{}

type JsonRpcHandler struct {
	registry map[string] JsonRpcCallback
}


func (j *JsonRpcHandler) Register(name string, fn JsonRpcCallback) {
	if j.registry == nil {
		j.registry = make(map[string] JsonRpcCallback)
	}
	
	j.registry[name] = fn
}

func (j *JsonRpcHandler) ExecuteJsonPayload(request map[string] interface{}) map[string] interface{} {
	method := request["method"]
	methodStr := method.(string)

	requestId := request["id"]

	params := request["params"].(map[string] interface{})

	methodValue := j.registry[methodStr]
	methodResult := methodValue(params)

	result := make(map[string] interface{})
	result["jsonrpc"] = "2.0"
	result["result"] = methodResult
	result["id"] = requestId

	return result
}

func makeJsonRpcError(code int, msg string) map[string] interface{} {
	error := make(map[string] interface{})
	
	error["code"] = code
	error["message"] = msg
	
	result := make(map[string] interface{})
	result["jsonrpc"] = "2.0"
	result["error"] = error
	result["id"] = nil
	
	return result
}

type SetStatusCodeFn func (code int) 

func (j *JsonRpcHandler) ExecuteJson(request io.Reader, response io.Writer, setStatus SetStatusCodeFn) {
	d := json.NewDecoder(request)

	var requestMap map[string] interface{}
	var responseObj map[string] interface{}
	
	err := d.Decode(&requestMap)
	if err == nil {
		responseObj = j.ExecuteJsonPayload(requestMap)
	} else {
		responseObj = makeJsonRpcError(-32700, "Parse error")
	}

	setStatus(http.StatusOK)

	e := json.NewEncoder(response)
	e.Encode(responseObj)
}

func (j JsonRpcHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 Only POST Allowed\n")
		return 
	}
	
	j.ExecuteJson(req.Body, w, func(status int) { 
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
	})

	return
}

func MewJsonRpcHandler (hub ThreadSafeServiceHub, timeline *Timeline) *JsonRpcHandler{

	r := new(JsonRpcHandler)
	
	r.Register("heartbeat", func(params map[string] interface{}) interface{} {
		name := params["name"].(string)
		
		err := hub.Heartbeat(name)

		if err == nil {
			return true
		}

		return makeJsonRpcError(100, err.String())
	})

	r.Register("log", func(params map[string] interface{}) interface{} {

		name := params["name"].(string)
		summary := params["summary"].(string)

		var severity int
		s := params["severity"]
		
		if _, ok := s.(float64) ; ok {
			severity = int(s.(float64))
		} else {
			severity = s.(int)
		}

// else {
//			severity = strconv.Atoi(s.(string))
//		}

		err := hub.Log(name, summary, severity, timeline.Now())
		
		if err == nil {
			return true
		}

		return makeJsonRpcError(100, err.String())
	})

	return r	
}

