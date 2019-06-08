GOPATH=$(shell pwd)/build
binaries := dvizz

all: golang bower

$(binaries):
	@echo Building $@
	GO111MODULE=on go build -o bin/$@ cmd/$@/main.go

build:
	@echo "🐳"
	docker build -t dvizz -f docker/Dockerfile .

fmt:
	find . -name '*.go' | grep -v vendor | grep -v build | xargs gofmt -w -s

test:
	go test ./cmd/... -race && go test ./internal/... -race

vet:
	go vet ./cmd/... && go vet ./internal/...

mock:
	mockgen -source internal/pkg/comms/server.go -destination internal/pkg/comms/mock_comms/mock_comms.go -package mock_comms

golang:
	mkdir -p $(GOPATH)
	export GOPATH=$(GOPATH) && go get -v -d
	mkdir -p dist
	export GOPATH=$(GOPATH) && go build -o dist/dvizz

bower: bower.json
	bower install

docker:
	docker build -f Dockerfile.dev .

.PHONY: $(binaries) build fmt