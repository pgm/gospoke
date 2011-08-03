#!/usr/bin/python
import urllib, json

f = urllib.urlopen("http://localhost:9199/jsonrpc", json.dumps({
  "jsonrpc": "2.0", "method": "log", "params": {"name": "Service Alpha", "severity": 4, "summary": "xxxxxx"}, "id": 1}))
print f.read()
