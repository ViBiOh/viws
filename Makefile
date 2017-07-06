default: deps fmt lint tst build

deps:
	go get -u github.com/golang/lint/golint

fmt:
	gofmt -s -w viws.go

lint:
	golint ./...
	go vet ./...

tst:
	go test ./...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo viws.go
