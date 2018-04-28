#!/bin/sh
sleep 5 # wait for redis
exec /go/bin/redis-proxy -redis "redis:6379"
