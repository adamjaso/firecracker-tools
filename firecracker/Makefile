SHELL := /bin/bash

FIRECRACKER_DIR ?= firecracker
FIRECRACKER_BIN = $(FIRECRACKER_DIR)/build/cargo_target/debug/firecracker
FIRECRACKER_URL = https://github.com/firecracker-microvm/firecracker

$(FIRECRACKER_DIR):
	git clone $(FIRECRACKER_URL) $(@)

$(FIRECRACKER_BIN): $(FIRECRACKER_DIR)
	pushd $(<) ; cargo build ; popd

build: $(FIRECRACKER_BIN)
	# no-op

clean:
	rm $(FIRECRACKER_BIN)

.PHONY: clean build
