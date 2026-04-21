# bm — Stremio-compatible CLI
# Usage: make | make build | make test

BINARY     ?= bm
CMD        := ./cmd/bm
BUILD_DIR  ?= bin
LDFLAGS    ?=

.PHONY: all build install test fmt vet tidy clean run help

all: build

help:
	@echo "Targets:"
	@echo "  make build    - go build -> $(BUILD_DIR)/$(BINARY)"
	@echo "  make install  - go install $(CMD)"
	@echo "  make test     - go test ./..."
	@echo "  make fmt      - gofmt -w ."
	@echo "  make vet      - go vet ./..."
	@echo "  make tidy     - go mod tidy"
	@echo "  make clean    - rm $(BUILD_DIR)/$(BINARY)"
	@echo "  make run      - go run $(CMD)"
	@echo "  make check    - fmt, vet, test"

build:
	@mkdir -p "$(BUILD_DIR)"
	go build -ldflags "$(LDFLAGS)" -o "$(BUILD_DIR)/$(BINARY)" "$(CMD)"

install:
	go install -ldflags "$(LDFLAGS)" "$(CMD)"

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt vet test

clean:
	rm -f "$(BUILD_DIR)/$(BINARY)"

run:
	go run "$(CMD)"
