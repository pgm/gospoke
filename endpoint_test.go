package main

import (
	. "launchpad.net/gocheck"
	"bytes"
)


func (s *S) TestJsonParse(c *C) {
	var rpc JsonRpcHandler 
	value := 1.0

	rpc.Register("add", func(params map[string] interface{}) interface{} {
		value += params["value"].(float64)
		
		return nil
	})
	
	request:="{\"jsonrpc\": \"2.0\", \"method\": \"add\", \"params\": {\"value\": 23}, \"id\": 3}"
	response := bytes.NewBufferString("")
	
	rpc.ExecuteJson(bytes.NewBufferString(request), response, func(code int) { c.Assert(code, Equals, 200) } )
	
	c.Assert(response.String(), Equals, "{\"id\":3,\"jsonrpc\":\"2.0\",\"result\":null}\n")
}
