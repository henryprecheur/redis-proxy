#!/bin/sh
sleep 10 # wait for redis & proxy
cd /go/src/github.com/henryprecheur/redis-proxy
exec go test -v .
