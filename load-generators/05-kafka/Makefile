.PHONY: build run test docker-build docker-push clean

APP_NAME := kafka-load-generator
VERSION := $(shell git describe --tags --always --dirty)
REGISTRY := ghcr.io/raibid-labs

build:
	go build -o bin/$(APP_NAME) ./cmd

run:
	go run ./cmd/main.go

test:
	go test -v ./...

docker-build:
	docker build -t $(REGISTRY)/$(APP_NAME):$(VERSION) .
	docker tag $(REGISTRY)/$(APP_NAME):$(VERSION) $(REGISTRY)/$(APP_NAME):latest

docker-push:
	docker push $(REGISTRY)/$(APP_NAME):$(VERSION)
	docker push $(REGISTRY)/$(APP_NAME):latest

clean:
	rm -rf bin/
	go clean
