SHELL := /bin/bash

.PHONY: bot
bot:
	go install -v ./cmd/bot

.PHONY: build
build: bot

.PHONY: test
test:
	go test -v ./...

.PHONY: check-format
check-format:
	./tools/check_gofmt.sh

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -o ./docker/bot ./cmd/bot
	cd docker && docker build -t dastanng/gitbot:latest . &&  cd -
	rm ./docker/bot
