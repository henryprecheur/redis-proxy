build:
	docker image pull golang:1.10-alpine
	docker image pull redis
	docker build . -t henryprecheur/redis-proxy

test: build
	docker-compose up
	docker-compose down

all: test
