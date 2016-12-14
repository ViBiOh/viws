default: vet build

vet:
	go vet ./...

server:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo server.go
