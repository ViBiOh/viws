default: deps fmt lint tst build

deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/ViBiOh/alcotest/alcotest

fmt:
	goimports -w viws.go
	gofmt -s -w viws.go

lint:
	golint ./...
	go vet ./...

tst:
	go test ./...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo viws.go
