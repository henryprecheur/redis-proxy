build:
	docker-compose build

test: build
	docker-compose up

all: test
