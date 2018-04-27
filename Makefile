DOCKER=docker
GO=go

docker-pull:
	${DOCKER} pull golang
	${DOCKER} pull redis
