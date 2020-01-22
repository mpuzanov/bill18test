GOBASE=$(shell pwd)
RELEASE_DIR=$(GOBASE)
APP=bill18test
PACKAGES := $(shell go list ./... | grep -v /vendor/)

.DEFAULT_GOAL = build 

lint:
	golangci-lint run	

build: lint
	go build -v .

run:
	go run .

image:
	docker build -t puzanovma/bill18test . 

rundock:
	docker run --rm -it -p 8091:8091 puzanovma/bill18test

release:
	rm -rf ${RELEASE_DIR}${APP}*
	GOOS=windows GOARCH=amd64 go build -o ${RELEASE_DIR}${APP}.exe main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ${RELEASE_DIR}${APP} main.go

.PHONY: build run release lint