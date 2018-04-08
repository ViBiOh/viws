default: api docker-api

api: deps go

go: format lint tst bench build

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w */*.go */*/*.go
	gofmt -s -w */*.go */*/*.go

lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/viws cmd/viws.go

start-api:
	go run cmd/viws.go \
		-directory `pwd`/example \
		-push /index.css

docker-deps:
	curl -s -o cacert.pem https://curl.haxx.se/ca/cacert.pem

docker-login:
	echo $(DOCKER_PASS) | docker login -u $(DOCKER_USER) --password-stdin

docker-api: docker-build-api docker-push-api

docker-build-api: docker-deps
	docker build -t $(DOCKER_USER)/viws .

docker-push-api: docker-login
	docker push $(DOCKER_USER)/viws

.PHONY: api go deps format lint tst bench build start-api docker-deps docker-login docker-api docker-build docker-push
