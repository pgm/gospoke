#!/usr/bin/python
import urllib, json

f = urllib.urlopen("http://localhost:9099/jsonrpc", json.dumps({
  "jsonrpc": "2.0", "method": "heartbeat", "params": {"name": "Service Alpha"}, "id": 1}))
print f.read()
