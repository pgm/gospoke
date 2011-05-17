#!/usr/bin/python
import urllib, json

f = urllib.urlopen("http://localhost:12345/jsonrpc", json.dumps({
  "jsonrpc": "2.0", "method": "heartbeat", "params": {"name": "serviceA"}, "id": 1}))
print f.read()
