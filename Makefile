default: go docker

go: deps dev

dev: format lint tst bench build

docker: docker-deps docker-build

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w *.go
	gofmt -s -w *.go

lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/viws viws.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -installsuffix nocgo -o bin/viws-arm viws.go

start-api:
	go run viws.go \
		-directory `pwd`

docker-deps:
	curl -s -o cacert.pem https://curl.haxx.se/ca/cacert.pem

docker-build:
	docker build -t ${DOCKER_USER}/viws .
	docker build -t ${DOCKER_USER}/viws:arm -f Dockerfile_arm .

docker-push:
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
	docker push ${DOCKER_USER}/viws
	docker push ${DOCKER_USER}/viws:arm
