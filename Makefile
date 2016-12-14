default: lint vet test build

lint:
	go get -u github.com/golang/lint/golint
	golint ./...

vet:
	go vet ./...

test:
	go test ./...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo server.go
