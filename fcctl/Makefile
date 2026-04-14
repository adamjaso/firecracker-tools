SHELL := /bin/bash

FCCTL_SRC = $(shell find ./ -name *.go)
FCCTL_BIN = fcctl

$(FCCTL_BIN): $(FCCTL_SRC)
	go build -o $(@)

build: $(FCCTL_BIN)
	# no-op

clean:
	rm $(FCCTL_BIN)

.PHONY: clean build
