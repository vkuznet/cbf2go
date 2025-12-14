# ===============================
# Project settings
# ===============================

APP_NAME := cbf2go
BIN_DIR  := bin

SERVER_BIN := $(BIN_DIR)/cbf_server
INGEST_BIN := $(BIN_DIR)/cbf_ingest
PNG_BIN := $(BIN_DIR)/cbf2png

GO := go
GOFLAGS := -trimpath

# ===============================
# Default target
# ===============================

.PHONY: all
all: build

# ===============================
# Build targets
# ===============================

.PHONY: build
build: server ingest png

.PHONY: server
server:
	@echo "==> Building cbf_server"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(SERVER_BIN) ./cmd/cbf_server

.PHONY: ingest
ingest:
	@echo "==> Building cbf_ingest"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(INGEST_BIN) ./cmd/cbf_ingest

.PHONY: png
png:
	@echo "==> Building cbf_png"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(PNG_BIN) ./cmd/cbf_png

# ===============================
# Cross-compilation
# ===============================

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: build-macos
build-macos:
	GOOS=darwin GOARCH=arm64 $(MAKE) build

# ===============================
# Dev helpers
# ===============================

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

.PHONY: run-server
run-server:
	$(GO) run ./cmd/cbf_server

.PHONY: run-ingest
run-ingest:
	$(GO) run ./cmd/cbf_ingest --file test.cbf

# ===============================
# Sanity checks
# ===============================

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

