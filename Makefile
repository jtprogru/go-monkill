.PHONY:
.SILENT:
.DEFAULT_GOAL := run

build:
	#go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/monkill ./cmd/monkill/main.go
	go mod download && go build -o ./.bin/monkill main.go

run: build
	./.bin/monkill

install:
	go mod install .
