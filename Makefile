SHELL := /bin/bash

.PHONY: bot
bot:
	go install -v ./cmd/bot

.PHONY: build
build: bot

.PHONY: test
test:
	go test -v ./...

.PHONY: gofmt-check
gofmt-check:
	./tools/gofmt_check.sh

.PHONY: govet-check
govet-check:
	./tools/govet_check.sh

.PHONY: golint-check
golint-check:
	go list ./... | grep -v "/test/" | xargs -L1 golint -set_exit_status

.PHONY: all-check
all-check: gofmt-check golint-check govet-check

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -o ./docker/bot ./cmd/bot
	cd docker && docker build -t dastanng/gitbot:latest . &&  cd -
	rm ./docker/bot
