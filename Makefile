SHELL := /bin/bash

.PHONY: bot
bot:
	go install -v ./cmd/bot

.PHONY: build
build: bot

.PHONY: test
test:
	go test -v ./...

.PHONY: check-gofmt
check-gofmt:
	./tools/check_gofmt.sh

# https://github.com/golang/lint
.PHONY: check-golint
check-golint:
	go list ./... | xargs -L1 golint -set_exit_status

.PHONY: check-govet
check-govet:
	./tools/check_govet.sh

.PHONY: check-all
check-all: check-gofmt check-golint check-govet

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -o ./docker/bot ./cmd/bot
	cd docker && docker build -t dastanng/gitbot:latest . &&  cd -
	rm ./docker/bot
