#!make
SHELL := /bin/bash
.SILENT:
.DEFAULT_GOAL := help

# Global vars
export SYS_GO=$(shell which go)
export SYS_GOFMT=$(shell which gofmt)

export BINARY_DIR=dist
export BINARY_NAME=go-monkill

.PHONY: run.cmd
## Run as go run cmd/app/main.go
run.cmd: main.go
	$(SYS_GO) run main.go

.PHONY: run.bin
## Run as binary
run.bin: build.bin
	./$(BINARY_DIR)/$(BINARY_NAME)

.PHONY: install-deps
## Install all requirements
install-deps: go.mod
	$(SYS_GO) mod tidy

.PHONY: build.bin
## Build bin file from go
build.bin: main.go
	$(SYS_GO) mod download && CGO_ENABLED=0 $(SYS_GO) build -o ./$(BINARY_DIR)/$(BINARY_NAME) main.go

.PHONY: fmt
## Run go fmt
fmt:
	$(SYS_GOFMT) -s -w .

.PHONY: vet
## Run go vet ./...
vet:
	$(SYS_GO) vet ./...

.PHONY: clean
## Clean all artifacts
clean:
	rm -rf $(BINARY_DIR)

.PHONY: test
## Run all test
test:
	go test --short -coverprofile=cover.out -v ./...
	make test.coverage

.PHONY: test.coverage
## Run test coverage
test.coverage:
	$(SYS_GO) tool cover -func=cover.out

.PHONY: lint
## Run golangci-lint
lint:
	golangci-lint -v run --out-format=colored-line-number

.PHONY: help
## Show this help message
help:
	@echo "$$(tput bold)Available rules:$$(tput sgr0)"
	@echo
	@sed -n -e "/^## / { \
		h; \
		s/.*//; \
		:doc" \
		-e "H; \
		n; \
		s/^## //; \
		t doc" \
		-e "s/:.*//; \
		G; \
		s/\\n## /---/; \
		s/\\n/ /g; \
		p; \
	}" ${MAKEFILE_LIST} \
	| LC_ALL='C' sort --ignore-case \
	| awk -F '---' \
		-v ncol=$$(tput cols) \
		-v indent=19 \
		-v col_on="$$(tput setaf 6)" \
		-v col_off="$$(tput sgr0)" \
	'{ \
		printf "%s%*s%s ", col_on, -indent, $$1, col_off; \
		n = split($$2, words, " "); \
		line_length = ncol - indent; \
		for (i = 1; i <= n; i++) { \
			line_length -= length(words[i]) + 1; \
			if (line_length <= 0) { \
				line_length = ncol - indent - length(words[i]) - 1; \
				printf "\n%*s ", -indent, " "; \
			} \
			printf "%s ", words[i]; \
		} \
		printf "\n"; \
	}' \
	| more $(shell test $(shell uname) == Darwin && echo '--no-init --raw-control-chars')
