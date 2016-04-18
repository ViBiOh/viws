default: server docker

server:
	CGO_ENABLED=0 GOGC=off go build -installsuffix -ldflags="-s -w" nocgo src/server.go

docker:
	docker build -t vibioh/http --rm .
