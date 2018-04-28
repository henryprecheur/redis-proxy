DOCKER=docker
GO=go

build:
	${DOCKER} build . -t henryprecheur/redis-proxy
