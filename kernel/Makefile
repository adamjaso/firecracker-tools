SHELL := /bin/bash

LINUX_VER ?= 6.19.11
LINUX_CONF ?= linux-$(LINUX_VER)-config
LINUX_DIR ?= linux-$(LINUX_VER)
LINUX_TAR = $(LINUX_DIR).tar.xz
LINUX_URL = https://cdn.kernel.org/pub/linux/kernel/v6.x/$(LINUX_TAR)

VMLINUX = linux-$(LINUX_VER)-vmlinux
VMLINUX_CONF = $(VMLINUX).config
ARTIFACTS = $(VMLINUX) $(VMLINUX_CONF)

$(LINUX_TAR):
	wget -O$(@) $(LINUX_URL)

$(LINUX_DIR): $(LINUX_TAR)
	mkdir -p $(@)
	tar --strip-components=1 -C $(@) -xJf $(<)

$(LINUX_DIR)/.config: $(LINUX_DIR)
	cp -av $(LINUX_CONF) $(@)
	make -C $(<) olddefconfig 

$(LINUX_DIR)/vmlinux: $(LINUX_DIR) $(LINUX_DIR)/.config
	make -C $(<) -j$(shell nproc) vmlinux

$(VMLINUX): $(LINUX_DIR)/vmlinux
	cp -a $(<) $(@)

$(VMLINUX_CONF): $(LINUX_DIR)/.config $(VMLINUX)
	cp -a $(<) $(@)

build: $(ARTIFACTS)
	# no-op

clean:
	rm -r $(ARTIFACTS)

.PHONY: build clean
