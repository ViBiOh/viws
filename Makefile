default: lint vet tst build

lint:
	go get -u github.com/golang/lint/golint
	golint ./...

vet:
	go vet ./...

tst:
	go test ./...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo server.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o health_check health/health.go
