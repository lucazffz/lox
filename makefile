.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint:fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

build: vet generate
	go build -o ../bin
.PHONY:build

run: generate
	go run . 
.PHONY:run 

generate:
	go generate
.PHONY:generate
