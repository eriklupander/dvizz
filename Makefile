GOPATH=$(shell pwd)/build

all: golang bower

golang:
	mkdir -p $(GOPATH)
	export GOPATH=$(GOPATH) && go get -d
	mkdir -p dist
	export GOPATH=$(GOPATH) && go build -o dist/dvizz

bower: bower.json
	bower install
